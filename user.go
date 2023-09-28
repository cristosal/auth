package auth

import (
	"time"

	"github.com/cristosal/pgxx"
)

const (
	PageSize = 25
)

type User struct {
	ID          pgxx.ID    `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Password    string     `json:"-"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	LastLogin   *time.Time `json:"last_login"`
	CreatedAt   *time.Time `json:"created_at"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) IsConfirmed() bool {
	return u.ConfirmedAt != nil
}

func (u *User) Confirm() {
	now := time.Now()
	u.ConfirmedAt = &now
}

func (u *User) VerifyPassword(pass string) bool {
	err := verifyHash(u.Password, pass)
	return err == nil
}

func (r *Service) PaginateUsers(page int, q string) (*pgxx.PaginationResults[User], error) {
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

func (r *Service) UserByID(id pgxx.ID) (*User, error) {
	var u User
	if err := pgxx.One(r.db, &u, "where id = $1", id); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Service) UserByEmail(email string) (*User, error) {
	var u User
	if err := pgxx.One(r.db, &u, "where email = $1", email); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Service) UpdateUser(u *User) error {
	return pgxx.Exec(r.db, "update users set name = $1, email = $2, phone = $3 where id = $4", u.Name, u.Email, u.Phone, u.ID)
}
