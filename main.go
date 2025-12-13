package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"taran1s.share/controllers"
	"taran1s.share/templates"
	"taran1s.share/views"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "layout.gohtml", "hello.gohtml"))))

	r.Get("/sign-up", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "layout.gohtml", "signup.gohtml"))))

	http.ListenAndServe(":3000", r)
}
