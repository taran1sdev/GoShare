package controllers

import (
	"fmt"
	"net/http"

	"taran1s.share/context"
	"taran1s.share/models"
)

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

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}

	UserService    *models.UserService
	SessionService *models.SessionService
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
		Email:        r.FormValue("email"),
		Forename:     r.FormValue("forename"),
		Surname:      r.FormValue("surname"),
		Password:     r.FormValue("password"),
		ConfirmPass:  r.FormValue("confirm"),
		InvalidEmail: false,
		NoMatch:      false,
		AuthFailed:   false,
	}

	user, err := u.UserService.Create(data)
	if err != nil {
		u.Templates.New.Execute(w, r, data)
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

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	data := &models.NewUser{}

	data.Email = r.FormValue("email")
	data.AuthFailed = false

	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := &models.NewUser{
		Email:      r.FormValue("email"),
		Password:   r.FormValue("password"),
		AuthFailed: true,
	}

	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		u.Templates.SignIn.Execute(w, r, data)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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
		fmt.Println(err)
		http.Error(w, "Something went wrong..", http.StatusInternalServerError)
		return
	}

	deleteCookie(w, CookieSession)

	http.Redirect(w, r, "/signin", http.StatusFound)
}
