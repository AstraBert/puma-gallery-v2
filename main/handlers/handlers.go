package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"puma-gallery/commons"
	"puma-gallery/db"
	"puma-gallery/templates"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

func HomeRoute(c *fiber.Ctx) error {
	client, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_API_KEY"), nil)
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.Page404("Impossible to connect to the database, sorry!").Render(c.Context(), c.Response().BodyWriter())
	}
	err = commons.AuthorizeGet(c)
	if err != nil {
		var results []commons.ImageData
		data, _, err := client.From("pictures").Select("*", "exact", false).Execute()
		if err != nil {
			c.Set("Content-Type", "text/html")
			return templates.Page404("There was an error while receiving the data").Render(c.Context(), c.Response().BodyWriter())
		}
		err = json.Unmarshal(data, &results)
		if err != nil {
			c.Set("Content-Type", "text/html")
			return templates.Page404("There was an error while receiving the data").Render(c.Context(), c.Response().BodyWriter())
		}
		c.Set("Content-Type", "text/html")
		return templates.MainPage(results, false).Render(c.Context(), c.Response().BodyWriter())
	} else {
		var results []commons.ImageData
		data, _, err := client.From("pictures").Select("*", "exact", false).Execute()
		if err != nil {
			c.Set("Content-Type", "text/html")
			return templates.Page404("There was an error while receiving the data").Render(c.Context(), c.Response().BodyWriter())
		}
		err = json.Unmarshal(data, &results)
		if err != nil {
			c.Set("Content-Type", "text/html")
			return templates.Page404("There was an error while receiving the data").Render(c.Context(), c.Response().BodyWriter())
		}
		c.Set("Content-Type", "text/html")
		return templates.MainPage(results, true).Render(c.Context(), c.Response().BodyWriter())
	}
}

func UploadRoute(c *fiber.Ctx) error {
	err := commons.AuthorizeGet(c)
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.AuthFailedPage().Render(c.Context(), c.Response().BodyWriter())
	} else {
		c.Set("Content-Type", "text/html")
		return templates.UploadImages(true).Render(c.Context(), c.Response().BodyWriter())
	}
}

func SinginRoute(c *fiber.Ctx) error {
	signin := templates.SignIn()
	c.Set("Content-Type", "text/html")
	return signin.Render(c.Context(), c.Response().BodyWriter())
}

func LoginUser(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	ctx := context.Background()
	sqlDb, err := commons.CreateNewDb()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error: " + err.Error()})
	}
	queries := db.New(sqlDb)
	user, err := queries.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			banners := templates.SingupBanner(errors.New("there is no user with this username"))
			return banners.Render(c.Context(), c.Response().BodyWriter())
		} else {
			banners := templates.SingupBanner(err)
			return banners.Render(c.Context(), c.Response().BodyWriter())
		}
	}
	if !commons.CompareHashToPassword(password, user.HashedPassword) {
		banners := templates.SingupBanner(errors.New("wrong username or password"))
		return banners.Render(c.Context(), c.Response().BodyWriter())
	} else {
		sess_token, errSes := commons.GenerateToken(32)
		csrf_token, errCsrf := commons.GenerateToken(32)
		if errSes != nil || errCsrf != nil {
			banners := templates.SingupBanner(errors.New("an error occurred while generating your authentication credentials"))
			return banners.Render(c.Context(), c.Response().BodyWriter())
		}
		err = queries.UpdateUserTokensLogin(ctx, db.UpdateUserTokensLoginParams{SessionToken: sql.NullString{String: sess_token, Valid: true}, CsrfToken: sql.NullString{String: csrf_token, Valid: true}, Username: username})
		if err != nil {
			banners := templates.SingupBanner(err)
			return banners.Render(c.Context(), c.Response().BodyWriter())
		} else {
			c.Cookie(&fiber.Cookie{
				Name:     "session_token",
				Value:    sess_token,
				Expires:  time.Now().Add(24 * time.Hour),
				HTTPOnly: true,
			})
			c.Cookie(&fiber.Cookie{
				Name:     "csrf_token",
				Value:    csrf_token,
				Expires:  time.Now().Add(24 * time.Hour),
				HTTPOnly: false,
			})
			c.Set("HX-Redirect", "/")
			return c.SendStatus(fiber.StatusOK)
		}
	}
}

func LogoutUser(c *fiber.Ctx) error {
	err := commons.AuthorizePost(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error: " + err.Error()})
	} else {
		ctx := context.Background()
		sqlDb, err := commons.CreateNewDb()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error: " + err.Error()})
		}
		queries := db.New(sqlDb)
		st := c.Cookies("session_token", "")
		csrf := c.Cookies("csrf_token", "")
		queries.UpdateUserTokensLogout(ctx, db.UpdateUserTokensLogoutParams{SessionToken: sql.NullString{String: st, Valid: true}, CsrfToken: sql.NullString{String: csrf, Valid: true}})
		c.Set("HX-Redirect", "/signin")
		return c.SendStatus(fiber.StatusOK)
	}
}

func PostPictures(c *fiber.Ctx) error {
	err := commons.AuthorizePost(c)
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	file, err := c.FormFile("image")
	caption := c.FormValue("caption", "")
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	src, err := file.Open()
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	defer src.Close()
	client, err := supabase.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_API_KEY"), nil)
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	relativePathUUID := uuid.Must(uuid.NewRandom())
	relativePath := "pumito/" + relativePathUUID.String() + "-" + file.Filename
	ctTp := "image/png"
	_, err = client.Storage.UploadFile(os.Getenv("SUPABASE_BUCKET_ID"), relativePath, src, storage_go.FileOptions{ContentType: &ctTp})
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	url := client.Storage.GetPublicUrl(os.Getenv("SUPABASE_BUCKET_ID"), relativePath)
	newImage := commons.ImageToUpload{Url: url.SignedURL, FilePath: relativePath, Caption: caption}
	_, _, err = client.From("pictures").Insert(newImage, true, "", "*", "exact").Execute()
	if err != nil {
		c.Set("Content-Type", "text/html")
		return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
	}
	c.Set("Content-Type", "text/html")
	return templates.SingupBanner(err).Render(c.Context(), c.Response().BodyWriter())
}
