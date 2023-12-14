package auth

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/cristosal/orm"
)

const (
	TokenDuration = time.Hour
)

var (
	ErrUserExists       = errors.New("user exists")
	ErrUserNotFound     = errors.New("not found")
	ErrNameRequired     = errors.New("name is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrInvalidToken     = errors.New("invalid token")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotFound    = errors.New("token not found")
)

type (
	RegistrationRequest struct {
		Name     string
		Email    string
		Phone    string
		Password string
	}

	RegistrationToken struct {
		UserID  int64
		Email   string
		Token   string
		Expires time.Time
	}

	RegistrationResponse struct {
		UserID int64
		Name   string
		Email  string
		Phone  string
		Token  string
	}
)

func (RegistrationToken) TableName() string {
	return "registration_tokens"
}

func (r *UserRepo) Register(req *RegistrationRequest) (*RegistrationResponse, error) {
	var (
		name  = req.Name
		email = req.Email
		phone = req.Phone
		pass  = req.Password
	)

	// sanitize values
	name = strings.Trim(name, " ")
	email = strings.ToLower(strings.Trim(email, " "))
	phone = strings.Trim(phone, " ")

	if name == "" {
		return nil, ErrNameRequired
	}

	if email == "" {
		return nil, ErrEmailRequired
	}

	if pass == "" {
		return nil, ErrPasswordRequired
	}

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

	res := RegistrationResponse{
		UserID: uid,
		Name:   name,
		Email:  email,
		Phone:  phone,
		Token:  tok,
	}

	return &res, nil
}

// ConfirmRegistration confirms a users account if a registration token is found matching tok
func (r *UserRepo) ConfirmRegistration(tok string) (*User, error) {
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
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrInvalidToken
		}

		return nil, err
	}

	if expires.Before(time.Now()) {
		// delete token
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

// RenewRegistration generates another registration token for the given user.
// Returns ErrTokenNotFound if a registration token was not available.
// To issue a renewal, a token must have already been generated
func (r *UserRepo) RenewRegistration(uid int64) (t *RegistrationToken, err error) {
	if err := orm.Get(r.db, t, "where user_id = $1", uid); err != nil {
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrTokenNotFound
		}

		return nil, err
	}
	tok, err := GenerateToken(16)
	if err != nil {
		return nil, err
	}

	t.Token = tok
	t.Expires = time.Now().Add(time.Hour * 3)

	if err := orm.Update(r.db, t, "where user_id = $1", uid); err != nil {
		return nil, err
	}

	return t, nil
}
