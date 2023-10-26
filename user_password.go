package auth

import (
	"context"
	"errors"
	"time"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 10

type PasswordReseter interface {
	ResetPassword(uid pgxx.ID, pass string) error
	RequestPasswordReset(email string) (token string, err error)
	ConfirmPasswordReset(token, pass string) error
}

func (r *UserPgxService) RequestPasswordReset(email string) (tok string, err error) {
	var (
		id   int64
		name string
		ctx  = context.Background()
	)

	// check if user exists.
	row := r.db.QueryRow(ctx, "select id, name from users where email = $1", email)
	if err = row.Scan(&id, &name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}

		return "", err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return
	}

	defer tx.Rollback(ctx)

	// delete existing password tokens
	_, err = tx.Exec(ctx, "delete from pass_tokens where user_id = $1", id)
	if err != nil {
		return
	}

	tok, err = GenerateToken(16)
	if err != nil {
		return
	}

	// insert token with expiry value
	_, err = tx.Exec(ctx, "insert into pass_tokens (user_id, token, expires) values ($1, $2, $3)", id, tok, time.Now().Add(time.Hour*3))
	if err != nil {
		return
	}

	err = tx.Commit(ctx)
	return
}

func (r *UserPgxService) ConfirmPasswordReset(token, pass string) error {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// get user assosciated with token
	row := tx.QueryRow(ctx, "select user_id from pass_tokens where token = $1", token)

	var uid pgxx.ID
	if err = row.Scan(&uid); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidToken
		}

		return err
	}

	// hash the new password
	newpass, err := PasswordHash(pass)
	if err != nil {
		return err
	}

	// update password
	_, err = tx.Exec(ctx, "update users set password = $1 where id = $2", newpass, uid)
	if err != nil {
		return err
	}

	// remove token
	_, err = tx.Exec(ctx, "delete from pass_tokens where user_id = $1 and token = $2", uid, token)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *UserPgxService) ResetPassword(uid pgxx.ID, pass string) error {
	ctx := context.Background()

	hashed, err := PasswordHash(pass)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, "update users set password = $1 where id = $2", hashed, uid)
	return err
}

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
