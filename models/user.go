package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// This type stores our user data after authentication
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

// This type is used for unauthenticated users during sign in / register
// it's used to pass data to the view and handle cases where the user
// submits invalid data
type NewUser struct {
	Email       string
	Forename    string
	Surname     string
	Password    string
	ConfirmPass string
}

// Helper functions for password hashing
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

// Helper to check email is valid format
func checkEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Register a new user
func (us *UserService) Create(nu *NewUser) (*User, error) {
	nu.Email = strings.ToLower(nu.Email)

	// Since we are using a pointer to the NewUser type
	// we can return an error and update the the type
	// and the view will render the error message

	if !checkEmail(nu.Email) {
		return nil, ErrInvalidEmail
	}

	if nu.Password != nu.ConfirmPass {
		return nil, ErrPasswordMatch
	}

	hash, err := getHashedPassword(nu.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        nu.Email,
		Forename:     nu.Forename,
		Surname:      nu.Surname,
		PasswordHash: hash,
	}

	row := us.DB.QueryRow(`
		INSERT INTO users(email, forename, surname, password_hash)
		VALUES ($1,$2,$3,$4) RETURNING id`, nu.Email, nu.Forename, nu.Surname, hash)

	err = row.Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailExists
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// Authenticate as an existing user
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
	if errors.Is(sql.ErrNoRows, err) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !checkPassword([]byte(user.PasswordHash), password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (us *UserService) UpdatePassword(userID int, password string) error {
	hashedBytes, err := getHashedPassword(password)
	if err != nil {
		return fmt.Errorf("updatePassword: %w", err)
	}

	passwordHash := string(hashedBytes)
	_, err = us.DB.Exec(`
		UPDATE users
		SET password_hash = $2
		WHERE id = $1;`, userID, passwordHash)

	if err != nil {
		return fmt.Errorf("updatePassword: %w", err)
	}

	return nil
}
