package auth

import "github.com/cristosal/pgxx"

type UserRepo interface {
	Paginate(page int, q string) (*pgxx.PaginationResults[User], error)
	ByID(pgxx.ID) (*User, error)
	ByEmail(string) (*User, error)
	UpdateInfo(*User) error
}

// Paginate paginates through users returning the most recent ones first
func (r *UserPgxService) Paginate(page int, q string) (*pgxx.PaginationResults[User], error) {
	return pgxx.Paginate[User](r.db, &pgxx.PaginationOptions{
		Record:        &User{},
		Query:         q,
		QueryColumns:  []string{"name", "email", "phone"},
		Page:          page,
		PageSize:      PageSize,
		SortBy:        "created_at",
		SortDirection: pgxx.SortDescending,
	})
}

// ByID returns a user by id field
func (r *UserPgxService) ByID(id pgxx.ID) (*User, error) {
	var u User
	if err := pgxx.One(r.db, &u, "where id = $1", id); err != nil {
		return nil, err
	}

	return &u, nil
}

// ByEmail returns a user by email
func (r *UserPgxService) ByEmail(email string) (*User, error) {
	var u User
	if err := pgxx.One(r.db, &u, "where email = $1", email); err != nil {
		return nil, err
	}

	return &u, nil
}

// UpdateInfo updates the users info, excluding the password
func (r *UserPgxService) UpdateInfo(u *User) error {
	return pgxx.Exec(r.db, "update users set name = $1, email = $2, phone = $3 where id = $4", u.Name, u.Email, u.Phone, u.ID)
}
