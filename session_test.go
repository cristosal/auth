package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/cristosal/auth"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v5"
)

func TestUserSessions(t *testing.T) {
	db, _ := pgx.Connect(context.Background(), os.Getenv("CONNECTION_STRING"))
	rd := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
	store := auth.NewSessionStore(db, rd)
	sess := auth.NewSession()
	sess.ExpiresAt = time.Now().Add(time.Minute)

	if err := store.DeleteUserSessions(1); err != nil {
		t.Fatal(err)
	}

	sess.User = &auth.User{
		ID:   1,
		Name: "Test User",
	}

	if err := store.Save(&sess); err != nil {
		t.Fatal(err)
	}

	userSessions, err := store.UserSessions(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(userSessions) != 1 {
		t.Fatal("expected 1 session")
	}
}

func TestSessionExpired(t *testing.T) {
	d := time.Second * 2
	sess := auth.NewSession()
	sess.ExpiresAt = time.Now().Add(d)
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
