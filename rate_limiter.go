package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

// MaxAttemptsError is returned by Limit implementations when the maximum number of attempts have been hit
type MaxAttemptsError struct {
	TTL      time.Duration // time until window becomes available
	Attempts int           // attempt limit that was reached
}

// As satisfies the As interface for use with errors.As
func (e *MaxAttemptsError) As(v interface{}) bool {
	if err, ok := v.(*MaxAttemptsError); ok {
		err.Attempts = e.Attempts
		err.TTL = e.TTL
		return true
	}

	return false
}

// Error implementation
func (err MaxAttemptsError) Error() string {
	return fmt.Sprintf("max attempts: ttl %s", err.TTL)
}

// Limiter defines the interface for rate limiting
type Limiter interface {
	// Limit increases an internal counter everytime it is called.
	// When the counter reaches the value specified by max within a given time window, MaxAttemptsError is returned.
	// The error is returned until the time duration specifed by window has elapsed.
	Limit(key string, max int, window time.Duration) error

	// Reset resets the internal counter to 0
	Reset(key string) error
}

// redisLimiter is the implementation for Limiter using redis as cache
type redisLimiter struct{ Client *redis.Client }

func NewRedisRateLimiter(cl *redis.Client) Limiter {
	return &redisLimiter{cl}
}

// Limit is the implementation of Limiter interface
func (l redisLimiter) Limit(key string, max int, window time.Duration) error {
	hits, err := l.get(key)
	if err != nil {
		return err
	}

	// see if we have hit the limit
	if hits >= max {
		// read ttl to send back to client
		ttl, err := l.ttl(key)
		if err != nil {
			return err
		}

		return MaxAttemptsError{ttl, hits}
	}

	// we hit the cache
	_, err = l.hit(key)
	if err != nil {
		return err
	}

	// and update the expiry window
	if err := l.expire(key, window); err != nil {
		return err
	}

	return nil
}

func (l redisLimiter) Reset(key string) error {
	cmd := l.Client.Del(key)
	return cmd.Err()
}

func (l redisLimiter) get(key string) (hits int, err error) {
	cmd := l.Client.Get(key)
	err = cmd.Err()

	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return strconv.Atoi(cmd.Val())
}

func (l redisLimiter) hit(key string) (hits int, err error) {
	cmd := l.Client.Incr(key)
	err = cmd.Err()
	hits = int(cmd.Val())
	return
}

func (l redisLimiter) expire(key string, duration time.Duration) error {
	cmd := l.Client.Expire(key, duration)
	return cmd.Err()
}

func (l redisLimiter) ttl(key string) (time.Duration, error) {
	return l.Client.TTL(key).Result()
}
