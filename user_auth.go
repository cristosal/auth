package auth

import (
	"database/sql"
	"errors"
)

func (r *UserRepo) Authenticate(email, pass string) (*User, error) {
	u, err := r.ByEmail(email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}

	if ok := u.VerifyPassword(pass); !ok {
		return nil, ErrUnauthorized
	}

	row := r.db.QueryRow("update users set last_login = now() where id = $1 returning last_login", u.ID)
	if err := row.Scan(&u.LastLogin); err != nil {
		return nil, err
	}

	return u, nil
}
