package controllers

import (
	"net/http"

	"taran1s.share/models"
)

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}

	UserService *models.UserService
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

	err = u.UserService.Create(data)
	if err != nil {
		u.Templates.New.Execute(w, r, data)
		return
	}

	u.Templates.SignIn.Execute(w, r, data)
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

	cookie := http.Cookie{
		Name:     "email",
		Value:    user.Email,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
}
