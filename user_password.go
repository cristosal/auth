package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cristosal/orm"
	"github.com/cristosal/orm/schema"
	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 10

type PasswordResetToken struct {
	UserID  int64
	Email   string
	Token   string
	Expires time.Time
}

type PasswordReset struct {
	Token    string
	Password string
}

func (PasswordResetToken) TableName() string {
	return "pass_tokens"
}

func (r *UserRepo) RequestPasswordReset(email string) (*PasswordResetToken, error) {
	var (
		id   int64
		name string
	)

	// check if user exists.
	row := r.db.QueryRow("select id, name from users where email = $1", email)
	if err := row.Scan(&id, &name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var t PasswordResetToken
	// delete existing password tokens for the given user
	if err := orm.Remove(tx, &t, "where user_id = $1", id); err != nil {
		return nil, err
	}

	token, err := GenerateToken(16)
	if err != nil {
		return nil, err
	}

	t.UserID = id
	t.Token = token
	t.Email = email
	t.Expires = time.Now().Add(time.Hour * 3)

	if err := orm.Add(tx, &t); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

// ConfirmPasswordReset
func (r *UserRepo) ConfirmPasswordReset(reset *PasswordReset) (*User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var (
		expires time.Time
		uid     int64
	)

	// get user assosciated with token
	row := tx.QueryRow("select user_id, expires from pass_tokens where token = $1", reset.Token)
	if err = row.Scan(&uid, &expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}

		return nil, err
	}

	if expires.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// hash the new password
	password, err := PasswordHash(reset.Password)
	if err != nil {
		return nil, err
	}

	var u User
	cols := schema.MustGet(&u).Fields.Columns().List()
	err = orm.QueryRow(tx, &u, fmt.Sprintf("update users set password = $1 where id = $2 returning %s", cols), password, uid)
	if err != nil {
		return nil, err
	}

	// remove token
	_, err = tx.Exec("delete from pass_tokens where user_id = $1 and token = $2", uid, reset.Token)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) ResetPassword(uid int64, pass string) error {
	hashed, err := PasswordHash(pass)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("update users set password = $1 where id = $2", hashed, uid)
	return err
}

// PasswordHash performs a bcrypt hash for the password based on PasswordHashCost
func PasswordHash(pass string) (string, error) {
	str, err := bcrypt.GenerateFromPassword([]byte(pass), PasswordHashCost)
	if err != nil {
		return "", err
	}

	return string(str), err
}

func verifyHash(hash string, pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}
