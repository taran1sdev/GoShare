package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"taran1s.share/controllers"
	"taran1s.share/models"
	"taran1s.share/templates"
	"taran1s.share/views"
)

func main() {
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "layout.gohtml", "hello.gohtml"))))

	userService := models.UserService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService: &userService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "signup.gohtml",
	))

	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "signin.gohtml",
	))

	r.Get("/signup", usersC.New)
	r.Get("/signin", usersC.SignIn)

	r.Post("/users", usersC.Create)
	r.Post("/signin", usersC.Authenticate)
	http.ListenAndServe(":3000", r)
}
