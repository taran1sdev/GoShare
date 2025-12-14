package models

import (
	"database/sql"
	"fmt"
	"net/mail"
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
	Email        string
	Forename     string
	Surname      string
	Password     string
	ConfirmPass  string
	InvalidEmail bool
	NoMatch      bool
	AuthFailed   bool
}

func getHashedPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashBytes), nil
}

func checkPassword(hash []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

func checkEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (us *UserService) Create(nu *NewUser) error {
	nu.Email = strings.ToLower(nu.Email)

	if !checkEmail(nu.Email) {
		nu.InvalidEmail = true
		return fmt.Errorf("invalid email address")
	}

	if nu.Password != nu.ConfirmPass {
		nu.NoMatch = true
		return fmt.Errorf("passwords do not match")
	}

	hash, err := getHashedPassword(nu.Password)
	if err != nil {
		return err
	}

	row := us.DB.QueryRow(`
		INSERT INTO users(email, forename, surname, password_hash)
		VALUES ($1,$2,$3,$4) RETURNING id`, nu.Email, nu.Forename, nu.Surname, hash)

	var id int
	err = row.Scan(&id)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)

	user := &User{
		Email: email,
	}

	row := us.DB.QueryRow(`
		SELECT id, password_hash, forename, surname 
		FROM users
		WHERE email=$1`, email)

	err := row.Scan(&user.ID, &user.PasswordHash, &user.Forename, &user.Surname)
	if err != nil {
		return nil, err
	}

	if !checkPassword([]byte(user.PasswordHash), password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}
