package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"taran1s.share/context"
	"taran1s.share/errors"
	"taran1s.share/models"
)

var ErrGeneric error = fmt.Errorf("Something went wrong... Please try again")

type UserMiddleware struct {
	SessionService *models.SessionService
}

// Set up middleware that reads the session cookie from the request
// and adds the User object associated with the session to the
// http request context
func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithUser(r.Context(), user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type Users struct {
	Templates struct {
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
	}

	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	data := &models.NewUser{}

	u.Templates.New.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := &models.NewUser{
		Email:       r.FormValue("email"),
		Forename:    r.FormValue("forename"),
		Surname:     r.FormValue("surname"),
		Password:    r.FormValue("password"),
		ConfirmPass: r.FormValue("confirm"),
	}

	user, err := u.UserService.Create(data)
	if err != nil {
		if errors.Is(models.ErrInvalidEmail, err) ||
			errors.Is(models.ErrPasswordMatch, err) ||
			errors.Is(models.ErrEmailExists, err) {
			err = errors.Public(err, err.Error())
		}
		u.Templates.New.Execute(w, r, data, err)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	data := &models.NewUser{}

	data.Email = r.FormValue("email")

	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	data := &models.NewUser{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	user, err := u.UserService.Authenticate(data.Email, data.Password)

	if err != nil {
		if errors.Is(models.ErrInvalidCredentials, err) {
			err = errors.Public(err, err.Error())
		}

		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	fmt.Fprintf(w, "Current user: %s", user.Email)
}

func (u Users) SignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	err = u.SessionService.Delete(token)
	if err != nil {
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
		return
	}

	deleteCookie(w, CookieSession)

	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")

	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	vals := url.Values{
		"token": {pwReset.Token},
	}

	// Make this configurable
	resetURL := "http://localhost:3000/reset-pw?" + vals.Encode()
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
		Error string
	}

	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not get the form values...", http.StatusBadRequest)
		return
	}

	var data struct {
		Token    string
		Password string
	}

	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		if errors.Is(models.ErrTokenExpired, err) {
			err = errors.Public(err, err.Error())
		}

		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}
