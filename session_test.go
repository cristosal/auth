package auth_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/cristosal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPgxSessionStore(t *testing.T) {
	db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		t.Fatal(err)
	}

	pgxStore := auth.NewSessionRepo(db)
	pgxStore.Drop()

	var (
		sess    = auth.NewSession(time.Now().Add(time.Minute))
		msgType = "success"
		msg     = "flash message"
	)
	sess.Flash(msgType, msg)

	if err := pgxStore.Save(&sess); err != nil {
		t.Fatal(err)
	}

	found, err := pgxStore.ByID(sess.ID)
	if err != nil {
		t.Fatal(err)
	}

	if found.Message != msg {
		t.Fatal("expected message to match")
	}

	if found.MessageType != msgType {
		t.Fatal("expected message types to match")
	}

}

func TestSessionExpired(t *testing.T) {
	d := time.Second * 2
	sess := auth.NewSession(time.Now().Add(d))
	if sess.Expired() {
		t.Fatal("expected session not to be expired")
	}
	time.Sleep(d)
	if !sess.Expired() {
		t.Fatal("expected session to be expired")
	}
}

func TestGenerateToken(t *testing.T) {

	tt := [][]int{
		{16, 32},
		{32, 64},
		{100, 200},
	}

	for _, tc := range tt {
		tok, err := auth.GenerateToken(tc[0])
		if err != nil {
			t.Fatal(err)
		}

		if len(tok) != tc[1] {
			t.Fatalf("expected %d to %d, got: %d", tc[0], tc[1], len(tok))
		}
	}
}
