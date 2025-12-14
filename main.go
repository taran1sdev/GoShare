package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"taran1s.share/controllers"
	"taran1s.share/migrations"
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

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		fmt.Println(err)
	}

	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "layout.gohtml", "hello.gohtml"))))

	userService := models.UserService{
		DB: db,
	}

	sessionService := models.SessionService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
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
	r.Post("/users", usersC.Create)

	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.Authenticate)

	r.Post("/signout", usersC.SignOut)

	r.Get("/users/me", usersC.CurrentUser)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "I think you got lost...", http.StatusNotFound)
	})

	// User middleware
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	// CSRF protection
	// Hard coded and insecure for now
	csrfKey := "gFvi45R4fy7xNVlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(false),
		csrf.TrustedOrigins([]string{"localhost:3000"}),
	)

	fmt.Println("Server starting on :3000")

	http.ListenAndServe(":3000", csrfMw(umw.SetUser(r)))
}
