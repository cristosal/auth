package auth

import (
	"errors"

	"github.com/cristosal/pgxx"
	"github.com/jackc/pgx/v5"
)

type Authenticator interface {
	Authenticate(email, pass string) (*User, error)
}

func (s *Service) Authenticate(email, pass string) (*User, error) {
	u, err := s.UserByEmail(email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnauthorized
	}

	if err != nil {
		return nil, err
	}

	if ok := u.VerifyPassword(pass); !ok {
		return nil, ErrUnauthorized
	}

	if err := pgxx.Exec(s.db, "update users set last_login = now() where id = $1", u.ID); err != nil {
		return nil, err
	}

	return u, nil
}
