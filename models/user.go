package models

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Forename     string
	Surname      string
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

type NewUser struct {
	Email    string
	Forename string
	Surname  string
	Password string
}

func (us *UserService) Create(nu *NewUser) (*User, error) {
	nu.Email = strings.ToLower(nu.Email)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hash := string(hashBytes)

	row := us.DB.QueryRow(`
		INSERT INTO users(email, forename, surname, password_hash)
		VALUES ($1,$2,$3,$4) RETURNING id`, nu.Email, nu.Forename, nu.Surname, hash)

	var id int
	err = row.Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &User{
		ID:           id,
		Forename:     nu.Forename,
		Surname:      nu.Surname,
		Email:        nu.Email,
		PasswordHash: hash,
	}, nil
}
