package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

var url = "localhost:6379"

func getLimiter() Limiter {
	rds := redis.NewClient(&redis.Options{
		Addr:       url,
		MaxRetries: 3,
	})

	limiter := NewRedisRateLimiter(rds)
	return limiter
}

func TestLimit1Passes(t *testing.T) {
	l := getLimiter()

	if err := l.Limit("test", 1, 0); err != nil {
		t.Fatal(err)
	}
}

func TestLimit2Fails(t *testing.T) {
	l := getLimiter()
	l.Limit("test", 1, time.Second)
	if err := l.Limit("test", 1, time.Second); err == nil {
		t.Fatal("expected err got nil")
	}

}

func TestAsMaxAttemptsError(t *testing.T) {
	l := getLimiter()
	err := l.Limit("test", 0, time.Second)

	var e MaxAttemptsError
	if !errors.As(err, &e) {
		t.Fatal("expected errors as to be true for max attempts error")
	}
}

func TestRateLimiter(t *testing.T) {
	l := getLimiter()

	n := 2
	for i := 0; i < n; i++ {
		l.Limit("test_key", n, time.Duration(time.Second))
	}

}
