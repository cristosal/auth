package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/cristosal/auth"
	"github.com/jackc/pgx/v5"
)

func NewTestService(t *testing.T) *auth.Service {
	conn, err := pgx.Connect(context.Background(), os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	return auth.New(conn)
}

func TestPermission(t *testing.T) {
	svc := NewTestService(t)
	t.Cleanup(func() {
		svc.ClearPermissions()
	})

	perms := []auth.Permission{
		{Name: "test"},
		{Name: "test1"},
		{Name: "test2"},
	}

	if err := svc.SeedPermissions(perms); err != nil {
		t.Fatal(err)
	}

	if perms[0].ID == 0 {
		t.Fatal("expected id to be set")
	}

	// two consecutives seeds should not error
	if err := svc.SeedPermissions(perms); err != nil {
		t.Fatal(err)
	}
}
