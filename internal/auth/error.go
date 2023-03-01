package auth

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrBadRequest          = errors.New("bad request")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrLoginIsAlreadyTaken = errors.New("login is already taken")
)
