package auth_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/cristosal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewTestService(t *testing.T) *auth.PgxService {
	conn, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	return auth.NewPgxService(conn)
}

func TestPermission(t *testing.T) {
	svc := NewTestService(t)
	t.Cleanup(func() {
		svc.Permissions().Clear()
	})

	svc.Users()

	perms := []auth.Permission{
		{Name: "test"},
		{Name: "test1"},
		{Name: "test2"},
	}

	if err := svc.Permissions().Seed(perms); err != nil {
		t.Fatal(err)
	}

	if perms[0].ID == 0 {
		t.Fatal("expected id to be set")
	}

	// two consecutives seeds should not error
	if err := svc.Permissions().Seed(perms); err != nil {
		t.Fatal(err)
	}
}
