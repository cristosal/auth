package auth_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/cristosal/auth"
	"github.com/go-redis/redis/v7"
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

func TestUserSessions(t *testing.T) {
	rd := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
	store := auth.NewRedisSessionStore(rd)
	sess := auth.NewSession(time.Now().Add(time.Minute))

	if err := store.DeleteByUserID(1); err != nil {
		t.Fatal(err)
	}

	sess.User = &auth.User{
		ID:   1,
		Name: "Test User",
	}

	if err := store.Save(&sess); err != nil {
		t.Fatal(err)
	}

	userSessions, err := store.ByUserID(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(userSessions) != 1 {
		t.Fatal("expected 1 session")
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
