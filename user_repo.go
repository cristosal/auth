package auth

import (
	"strings"

	"github.com/cristosal/orm"
)

// Paginate paginates through users returning the most recent ones first
func (r *UserRepo) Paginate(page int, q string) ([]User, *orm.PaginationResults, error) {
	var users []User
	results, err := orm.Paginate(r.db, &users, &orm.PaginationOptions{
		Query:         q,
		QueryColumns:  []string{"name", "email", "phone"},
		Page:          page,
		PageSize:      PageSize,
		SortBy:        "created_at",
		SortDirection: "DESC",
	})
	if err != nil {
		return nil, nil, err
	}

	return users, results, nil
}

// ByID returns a user by id field
func (r *UserRepo) ByID(id int64) (*User, error) {
	var u User
	if err := orm.Get(r.db, &u, "where id = $1", id); err != nil {
		return nil, err
	}

	return &u, nil
}

func (UserRepo) sanitizeEmail(email string) string {
	return strings.ToLower(strings.Trim(email, " "))
}

// ByEmail returns a user by email
func (r *UserRepo) ByEmail(email string) (*User, error) {
	var u User
	if err := orm.Get(r.db, &u, "where email = $1", r.sanitizeEmail(email)); err != nil {
		return nil, err
	}

	return &u, nil
}

// UpdateInfo updates the users info, excluding the password
func (r *UserRepo) UpdateInfo(u *User) error {
	return orm.Exec(r.db, "update users set name = $1, email = $2, phone = $3 where id = $4", u.Name, u.Email, u.Phone, u.ID)
}
