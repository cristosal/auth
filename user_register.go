package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/cristosal/orm"
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
		UserID int64
		Name   string
		Email  string
		Phone  string
		Token  string
	}

	Registrator interface {
		Register(name, username, pass, phone string) (*Registration, error)
		ConfirmRegistration(tok string) (*User, error)
		RenewRegistration(uid int64) (tok string, err error)
	}
)

func (r *UserService) Register(name, email, pass, phone string) (*Registration, error) {

	// sanitize values
	name = strings.Trim(name, " ")
	email = strings.ToLower(strings.Trim(email, " "))
	phone = strings.Trim(phone, " ")

	row := r.db.QueryRow("select email from users where email = $1", email)

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

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	row = tx.QueryRow("insert into users (name, email, password, phone) values ($1, $2, $3, $4) returning id", name, email, newpass, phone)

	var uid int64
	if err = row.Scan(&uid); err != nil {
		return nil, err
	}

	_, err = tx.Exec("insert into registration_tokens (user_id, token, expires) values ($1, $2, $3)", uid, tok, time.Now().Add(TokenDuration))
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
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

func (r *UserService) ConfirmRegistration(tok string) (*User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var (
		uid     int64
		expires time.Time
		row     = tx.QueryRow("select user_id, expires from registration_tokens where token = $1", tok)
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

	_, err = tx.Exec("update users set confirmed_at = $1 where id = $2", time.Now(), uid)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("delete from registration_tokens where user_id = $1", uid)
	if err != nil {
		return nil, err
	}

	var u User
	if err := orm.Get(tx, &u, "where id = $1", uid); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserService) RenewRegistration(uid int64) (tok string, err error) {
	if err = orm.Exec(r.db, "select 1 from users where id = $1", uid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrUserNotFound
		}
		return
	}

	tx, err := r.db.Begin()
	if err != nil {
		return
	}

	defer tx.Rollback()

	if _, err = tx.Exec("delete from registration_tokens where user_id = $1", uid); err != nil {
		return
	}

	tok, err = GenerateToken(16)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec("insert into registration_tokens (user_id, token, expires) values ($1, $2, $3)", uid, tok, time.Now().Add(TokenDuration))
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return tok, nil
}
