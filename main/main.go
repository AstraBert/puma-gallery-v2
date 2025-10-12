package main

import (
	"log"

	"puma-gallery/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create a new Fiber app
	app := Setup()

	// Start the Fiber server on port 8000
	if err := app.Listen(":8000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func Setup() *fiber.App {
	// Create a new Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})

	app.Get("/", handlers.HomeRoute)
	app.Post("/login", handlers.LoginUser)
	app.Post("/logout", handlers.LogoutUser)
	app.Get("/signin", handlers.SinginRoute)
	app.Get("/pictures", handlers.UploadRoute)
	app.Post("/pictures", handlers.PostPictures)
	return app
}
