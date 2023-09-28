package auth_test

import (
	"testing"
	"time"

	"github.com/cristosal/auth"
)

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
