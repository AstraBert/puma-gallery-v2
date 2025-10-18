package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestHomeRoute(t *testing.T) {
	_, okApi := os.LookupEnv("SUPABASE_API_KEY")
	_, okUrl := os.LookupEnv("SUPABASE_URL")
	if !okApi || !okUrl {
		t.Skip("No SUPABASE_API_KEY or SUPABASE_URL available")
	}
	app := Setup()
	req, _ := http.NewRequest("GET", "/", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the home route, got: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Expected no error while reading the response body from the home route, got: %s", err.Error())
	}
	if strings.Contains(string(body), "There was an error") {
		t.Error("No error was expected on the server side while loading the home page, got one")
	}
}

func TestSigninRoute(t *testing.T) {
	app := Setup()
	req, _ := http.NewRequest("GET", "/signin", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the signin route, got: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Expected no error while reading the response body from the signin route, got: %s", err.Error())
	}
	if !strings.Contains(string(body), "Sign In to Pumito Gallery") {
		t.Errorf("No error was expected on the server side while loading the signin page, got one, got: %s", string(body))
	}
}

func TestPicturesRouteUnauthorized(t *testing.T) {
	app := Setup()
	req, _ := http.NewRequest("GET", "/pictures", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the pictures route, got: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Expected no error while reading the response body from the pictures route, got: %s", err.Error())
	}
	if !strings.Contains(string(body), "401") {
		t.Errorf("Unauthorized was expected to be returned when hitting the pictures route without the appropriate setup, got: %s", string(body))
	}
}

func TestLoginRouteError(t *testing.T) {
	app := Setup()
	formData := url.Values{}
	formData.Set("login", "myusername")
	formData.Set("password", "mypassword")
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the login route, got: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Expected no error while reading the response body from the login route, got: %s", err.Error())
	}
	if !strings.Contains(string(body), "Error!") {
		t.Errorf("An error was expected to be returned when hitting the login route without the appropriate setup, got %s", string(body))
	}
}

func TestLogoutRouteError(t *testing.T) {
	app := Setup()
	req, _ := http.NewRequest("POST", "/logout", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the logout route, got: %s", err.Error())
	}
	if res.StatusCode != 500 {
		t.Errorf("Expected the logout route to return an error without the appropriate setup, got status code: %d", res.StatusCode)
	}
}

func TestPicturesPostRouteUnauthorized(t *testing.T) {
	app := Setup()
	req, _ := http.NewRequest("POST", "/pictures", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Errorf("Expected no error while testing the pictures (post) route, got: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Expected no error while reading the response body from the pictures (post) route, got: %s", err.Error())
	}
	if !strings.Contains(string(body), "unauthorized") {
		t.Errorf("An error was expected to be returned when hitting the pictures (post) route without the appropriate setup, got %s", string(body))
	}
}
