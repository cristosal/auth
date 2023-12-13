package auth

import "github.com/cristosal/orm"

// Paginate paginates through users returning the most recent ones first
func (r *UserService) Paginate(page int, q string) ([]User, *orm.PaginationResults, error) {
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
func (r *UserService) ByID(id int64) (*User, error) {
	var u User
	if err := orm.Get(r.db, &u, "where id = $1", id); err != nil {
		return nil, err
	}

	return &u, nil
}

// ByEmail returns a user by email
func (r *UserService) ByEmail(email string) (*User, error) {
	var u User
	if err := orm.Get(r.db, &u, "where email = $1", email); err != nil {
		return nil, err
	}

	return &u, nil
}

// UpdateInfo updates the users info, excluding the password
func (r *UserService) UpdateInfo(u *User) error {
	return orm.Exec(r.db, "update users set name = $1, email = $2, phone = $3 where id = $4", u.Name, u.Email, u.Phone, u.ID)
}
