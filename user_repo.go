package auth

import (
	"errors"
	"fmt"
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
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &u, nil
}

// ByGroup returns a slice of users belonging to a group
func (r *UserRepo) ByGroup(gid int64) ([]User, error) {
	var u User
	cols := orm.Columns(u).PrefixedList("u")
	sql := fmt.Sprintf("select %s from %s u inner join group_users gu on gu.user_id = u.id where gu.group_id = $1", cols, u.TableName())
	var users []User
	if err := orm.Query(r.db, &users, sql, gid); err != nil {
		return nil, err
	}
	return users, nil
}

// ByEmail returns a user by email
func (r *UserRepo) ByEmail(email string) (*User, error) {
	var u User
	if err := orm.Get(r.db, &u, "where email = $1", r.SanitizeEmail(email)); err != nil {
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &u, nil
}

// UpdateInfo updates the users info, excluding the password
func (r *UserRepo) UpdateInfo(u *User) error {
	return orm.Exec(r.db, "update users set name = $1, email = $2, phone = $3 where id = $4", u.Name, u.Email, u.Phone, u.ID)
}

func (UserRepo) SanitizeEmail(email string) string {
	return strings.ToLower(strings.Trim(email, " "))
}
