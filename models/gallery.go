package models

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrGalleryNoExist error = fmt.Errorf("Gallery does not exist..")

type Gallery struct {
	ID     int
	UserID int
	Title  string
}

type GalleryService struct {
	DB *sql.DB
}

func (service *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userID,
	}

	row := service.DB.QueryRow(`
		INSERT INTO galleries (title, user_id)
		VALUES ($1,$2) RETURNING id;`, gallery.Title, gallery.UserID)

	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	return &gallery, nil
}

func (service *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}

	row := service.DB.QueryRow(`
		SELECT title, user_id 
		FROM galleries
		WHERE id = $1;`, id)

	err := row.Scan(&gallery.Title, &gallery.UserID)
	if err != nil {
		return nil, ErrGalleryNoExist
	}

	return &gallery, nil
}

func (service *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	rows, err := service.DB.Query(`
		SELECT id, title
		FROM galleries
		WHERE user_id = $1;`, userID)

	if errors.Is(sql.ErrNoRows, err) {
		return nil, ErrGalleryNoExist
	} else if err != nil {
		return nil, fmt.Errorf("byuserid: %w", err)
	}

	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}

		err := rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("byuserid: %w", err)
		}

		galleries = append(galleries, gallery)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("byuserid: %w", err)
	}

	return galleries, nil
}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(`
		UPDATE galleries
		SET title = $2
		WHERE id = $1;`, gallery.ID, gallery.Title)

	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

func (service *GalleryService) DeleteID(id int) error {
	_, err := service.DB.Exec(`
		DELETE FROM galleries
		WHERE id = $1;`, id)

	if err != nil {
		return fmt.Errorf("deleteid: %w", err)
	}

	return nil
}

func (service *GalleryService) DeleteUser(userID int) error {
	_, err := service.DB.Exec(`
		DELETE FROM galleries
		WHERE user_id = $1;`, userID)

	if err != nil {
		return fmt.Errorf("deleteuser: %w", err)
	}

	return nil
}
