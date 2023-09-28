package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
)

const (
	TokenDuration = time.Hour
)

var (
	ErrUserExists   = errors.New("user exists")
	ErrUserNotFound = errors.New("not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrUnauthorized = errors.New("unauthorized")
	ErrTokenExpired = errors.New("token expired")
)

type (
	Registration struct {
		UserID pgxx.ID
		Name   string
		Email  string
		Phone  string
		Token  string
	}

	Registrator interface {
		Register(name, username, pass, phone string) (*Registration, error)
		ConfirmRegistration(tok string) (*User, error)
		RenewRegistration(uid pgxx.ID) (tok string, err error)
	}
)

func (s *Service) Register(name, email, pass, phone string) (*Registration, error) {
	ctx := context.Background()

	// sanitize values
	name = strings.Trim(name, " ")
	email = strings.ToLower(strings.Trim(email, " "))
	phone = strings.Trim(phone, " ")

	row := s.db.QueryRow(ctx, "select email from users where email = $1", email)

	var found string

	// ignore error
	row.Scan(&found)
	if found != "" {
		return nil, ErrUserExists
	}

	newpass, err := PasswordHash(pass)
	if err != nil {
		return nil, err
	}

	tok, err := GenerateToken(16)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	row = tx.QueryRow(ctx, "insert into users (name, email, password, phone) values ($1, $2, $3, $4) returning id", name, email, newpass, phone)

	var uid pgxx.ID
	if err = row.Scan(&uid); err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, "insert into registration_tokens (user_id, token, expires) values ($1, $2, $3)", uid, tok, time.Now().Add(TokenDuration))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	res := Registration{
		UserID: uid,
		Name:   name,
		Email:  email,
		Phone:  phone,
		Token:  tok,
	}

	return &res, nil
}

func (s *Service) ConfirmRegistration(tok string) (*User, error) {
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	var (
		uid     pgxx.ID
		expires time.Time
		row     = tx.QueryRow(ctx, "select user_id, expires from registration_tokens where token = $1", tok)
	)

	if err = row.Scan(&uid, &expires); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrInvalidToken
		}
		return nil, err
	}

	if expires.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	_, err = tx.Exec(ctx, "update users set confirmed_at = $1 where id = $2", time.Now(), uid)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, "delete from registration_tokens where user_id = $1", uid)
	if err != nil {
		return nil, err
	}

	var u User
	if err := pgxx.One(tx, &u, "where id = $1", uid); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *Service) RenewRegistration(uid pgxx.ID) (tok string, err error) {
	if err = pgxx.Exec(s.db, "select 1 from users where id = $1", uid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrUserNotFound
		}
		return
	}

	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return
	}

	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, "delete from registration_tokens where user_id = $1", uid); err != nil {
		return
	}

	tok, err = GenerateToken(16)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, "insert into registration_tokens (user_id, token, expires) values ($1, $2, $3)", uid, tok, time.Now().Add(TokenDuration))
	if err != nil {
		return "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return "", err
	}

	return tok, nil
}
