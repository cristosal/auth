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

type UserPgxService struct{ db pgxx.DB }

func NewUserPgxService(db pgxx.DB) *UserPgxService {
	return &UserPgxService{db}
}

type UserService interface {
	UserRepo
	Authenticator
	PasswordReseter
}
