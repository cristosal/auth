package auth

import "errors"

// package wide errors go here
var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrEmailRequired      = errors.New("email is required")
	ErrInvalidToken       = errors.New("invalid token")
	ErrNameRequired       = errors.New("name is required")
	ErrPasswordRequired   = errors.New("password is required")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUserExists         = errors.New("user exists")
	ErrUserNotFound       = errors.New("user not found")
)
