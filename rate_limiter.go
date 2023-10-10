package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

var (
	ErrLimitReached = errors.New("limit reached")
)

// Limiter is the interface implemented by rate limiters
type Limiter interface {
	// Limit increases the hit count for a given key.
	// When count value reaches max within window duration, MaxAttemptsError is returned until window has elapsed.
	Limit(key string, max int, window time.Duration) error

	// TTL returns the time to live for the limit block.
	// Returns 0 duration if limit has not yet been reached.
	TTL(key string, max int) (ttl time.Duration)

	// Reset resets the limit for a given key.
	Reset(key string) error
}

// RedisLimiter is the implementation for Limiter using redis as cache
type RedisLimiter struct{ cl *redis.Client }

// NewRedisLimiter returns a Limiter implementation using redis as the underlying cache store
func NewRedisLimiter(cl *redis.Client) *RedisLimiter {
	return &RedisLimiter{cl}
}

// Limit is the implementation of Limiter interface.
// It returns ErrLimitReached when attempts have been exceeded.
func (l *RedisLimiter) Limit(key string, max int, window time.Duration) error {
	if l.limitReached(key, max) {
		return ErrLimitReached
	}

	// we hit the cache
	_, err := l.hit(key)
	if err != nil {
		return err
	}

	// and update the expiry window
	if err := l.expire(key, window); err != nil {
		return err
	}

	return nil
}

func (l *RedisLimiter) TTL(key string, max int) (ttl time.Duration) {
	if !l.limitReached(key, max) {
		return 0
	}

	ttl, err := l.ttl(key)
	if err != nil {
		return 0
	}

	return ttl
}

func (l *RedisLimiter) Reset(key string) error {
	cmd := l.cl.Del(key)
	return cmd.Err()
}

func (l *RedisLimiter) limitReached(key string, max int) bool {
	h, err := l.get(key)
	if err != nil {
		return false
	}

	return h >= max
}

func (l *RedisLimiter) get(key string) (hits int, err error) {
	cmd := l.cl.Get(key)
	err = cmd.Err()

	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return strconv.Atoi(cmd.Val())
}

func (l *RedisLimiter) hit(key string) (hits int, err error) {
	cmd := l.cl.Incr(key)
	err = cmd.Err()
	hits = int(cmd.Val())
	return
}

func (l *RedisLimiter) expire(key string, duration time.Duration) error {
	cmd := l.cl.Expire(key, duration)
	return cmd.Err()
}

func (l *RedisLimiter) ttl(key string) (time.Duration, error) {
	return l.cl.TTL(key).Result()
}
