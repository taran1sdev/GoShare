package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"taran1s.share/controllers"
	"taran1s.share/migrations"
	"taran1s.share/models"
	"taran1s.share/templates"
	"taran1s.share/views"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		Database: os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL"),
	}

	cfg.SMTP = models.SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
	}

	cfg.SMTP.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return cfg, err
	}

	secure := false
	if os.Getenv("CSRF_SECURE") != "" {
		secure = true
	}

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = secure

	cfg.Server.Address = fmt.Sprintf("%s:%s", os.Getenv("SERVER_ADDR"), os.Getenv("SERVER_PORT"))

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		fmt.Println(err)
	}

	userService := &models.UserService{
		DB: db,
	}

	sessionService := &models.SessionService{
		DB: db,
	}

	passwordResetService := &models.PasswordResetService{
		DB:            db,
		BytesPerToken: 32,
		Duration:      models.DefaultResetDuration,
	}

	emailService := models.NewEmailService(cfg.SMTP)

	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: passwordResetService,
		EmailService:         emailService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "signup.gohtml",
	))

	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "signin.gohtml",
	))

	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "forgotpw.gohtml",
	))

	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "checkemail.gohtml",
	))

	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "resetpw.gohtml",
	))

	galleryService := &models.GalleryService{
		DB: db,
	}

	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}

	galleriesC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "newgallery.gohtml",
	))

	galleriesC.Templates.Edit = views.Must(views.ParseFS(
		templates.FS,
		"layout.gohtml", "editgallery.gohtml",
	))

	// User middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// CSRF protection
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.TrustedOrigins([]string{"localhost:3000"}),
		csrf.Path("/"),
	)

	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(
		views.Must(views.ParseFS(templates.FS, "layout.gohtml", "hello.gohtml"))))

	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)

	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.Authenticate)

	r.Post("/signout", usersC.SignOut)

	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)

	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Get("/users/me", usersC.CurrentUser)

	r.Route("/galleries", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "I think you got lost...", http.StatusNotFound)
	})

	fmt.Println("Server starting on :3000")

	http.ListenAndServe(cfg.Server.Address, csrfMw(umw.SetUser(r)))
}
