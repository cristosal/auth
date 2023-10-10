package auth

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

var redisAddr = "localhost:6379"

func getLimiter() Limiter {
	rds := redis.NewClient(&redis.Options{
		Addr:       redisAddr,
		MaxRetries: 3,
	})

	limiter := NewRedisLimiter(rds)
	return limiter
}

func TestTTL(t *testing.T) {
	var (
		l   = getLimiter()
		max = 3
		k   = "test"
	)

	l.Limit(k, max, time.Minute)
	l.Limit(k, max, time.Minute)
	l.Limit(k, max, time.Minute)
	ttl := l.TTL(k, max)
	if ttl != time.Minute {
		t.Fatal("expected ttl to be minute")
	}
}

func TestLimitExpires(t *testing.T) {
	l := getLimiter()
	k := "test-limit-expires"
	// max of 2 means 2 hits are allowed within the window
	if err := l.Limit(k, 2, time.Second); err != nil {
		t.Fatal(err)
	}

	if err := l.Limit(k, 2, time.Second); err != nil {
		t.Fatal(err)
	}

	// we expect an error here
	if err := l.Limit(k, 2, time.Second); err == nil {
		t.Fatal("expected error got nil")
	}

	time.Sleep(time.Second)

	if err := l.Limit(k, 2, time.Second); err != nil {
		t.Fatal(err)
	}
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

func TestRateLimiter(t *testing.T) {
	l := getLimiter()

	n := 2
	for i := 0; i < n; i++ {
		l.Limit("test_key", n, time.Duration(time.Second))
	}

}
