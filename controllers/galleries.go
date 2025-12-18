package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"taran1s.share/context"
	"taran1s.share/errors"
	"taran1s.share/models"
)

type Galleries struct {
	Templates struct {
		New   Template
		Edit  Template
		Index Template
	}
	GalleryService *models.GalleryService
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
		Error string
	}

	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID int
		Title  string
	}

	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}

	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return
	}

	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}

	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return
	}

	title := r.FormValue("title")
	gallery.Title = title
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
	}

	var data struct {
		Galleries []Gallery
	}

	g.Templates.Index.Execute(w, r, data)
}
