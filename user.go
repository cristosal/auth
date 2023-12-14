package auth

import (
	"time"

	"github.com/cristosal/orm"
)

const (
	PageSize = 25
)

type User struct {
	ID          int64
	Name        string
	Email       string
	Phone       string
	Password    string `json:"-"`
	ConfirmedAt *time.Time
	LastLogin   *time.Time
	CreatedAt   *time.Time
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

type UserRepo struct{ db orm.DB }

func NewUserRepo(db orm.DB) *UserRepo {
	return &UserRepo{db}
}
