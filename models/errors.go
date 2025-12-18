package models

import "errors"

var (
	ErrInvalidCredentials = errors.New("Invalid Credentials")
	ErrInvalidEmail       = errors.New("Invalid Email Address")
	ErrEmailExists        = errors.New("There is already an account registered  with that email address")
	ErrPasswordMatch      = errors.New("Passwords do not match")
	ErrNotFound           = errors.New("Resource could not be found")
)
