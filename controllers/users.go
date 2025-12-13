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
	u.Templates.New.Execute(w, nil)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := &models.NewUser{
		Email:    r.FormValue("email"),
		Forename: r.FormValue("forename"),
		Surname:  r.FormValue("surname"),
		Password: r.FormValue("password"),
	}

	_, err = u.UserService.Create(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u.Templates.SignIn.Execute(w, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, data)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	_, err = u.UserService.Authenticate(email, password)
	if err != nil {
		u.Templates.SignIn.Execute(w, nil)
	}
}
