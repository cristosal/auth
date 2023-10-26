package auth_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/cristosal/auth"
	"github.com/jackc/pgx/v5"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "delete from users where email = $1", "pepito@gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	s := auth.NewPgxService(conn)

	// test sanitizes value
	reg, err := s.Users().Register("pepe   ", " pepito@gmail.com  ", "hello world", "123")
	if err != nil {
		t.Fatalf("unable to register %v", err)
	}

	_, err = s.Users().Register("pepe   ", "    pepito@gmail.com  ", "hello world", "123")
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
