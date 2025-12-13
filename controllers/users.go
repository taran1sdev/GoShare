package controllers

import (
	"fmt"
	"net/http"

	"taran1s.share/models"
)

type Users struct {
	Templates struct {
		New Template
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

	user, err := u.UserService.Create(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, "ID: ", user.ID, "email: ", user.Email, "forename: ", user.Forename)
}
