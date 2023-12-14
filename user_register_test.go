package auth_test

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/cristosal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRegister(t *testing.T) {
	// the problem here is we should use sqlmock
	conn, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Exec("delete from users where email = $1", "pepito@gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	s := auth.NewService(conn)
	req := &auth.RegistrationRequest{
		Name:     "pepe       ",
		Email:    " pepito@gmail.com   ",
		Phone:    "hello world",
		Password: "  123 ",
	}

	reg, err := s.Users().Register(req)
	if err != nil {
		t.Fatalf("unable to register %v", err)
	}

	_, err = s.Users().Register(req)
	if !errors.Is(err, auth.ErrUserExists) {
		t.Fatalf("expected user exists got %v", err)
	}

	if reg.Name != "pepe" {
		t.Fatal("expected to name to be sanitized")
	}

	if reg.Email != "pepito@gmail.com" {
		t.Fatal("expected email to be sanitized")
	}

	if reg.Token == "" {
		t.Fatal("expected token to be present")
	}

	_, err = s.Users().ConfirmRegistration("fail")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("expected invalid token got %v", err)
	}

	_, err = s.Users().ConfirmRegistration(reg.Token)
	if err != nil {
		t.Fatal("expected confirmation of token to not return anything")
	}

}
