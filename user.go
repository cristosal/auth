package auth

import (
	"time"

	"github.com/cristosal/orm"
)

const (
	PageSize = 25
)

type User struct {
	ID          int64      `json:"id"`
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

type UserService struct{ db orm.DB }

func NewUserService(db orm.DB) *UserService {
	return &UserService{db}
}
