package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

// MaxAttemptsError should be returned by Limit implementations when the maximum number of attempts have been hit
type MaxAttemptsError struct {
	TTL      time.Duration
	Attempts int
}

func (err MaxAttemptsError) Error() string {
	return fmt.Sprintf("max attempts: ttl %s", err.TTL)
}

// Limiter defines the interface for rate limiting
type Limiter interface {
	Limit(key string, maxAttempts int, expiration time.Duration) error
	Reset(key string) error
}

// RedisLimiter is the implementation for Limiter using redis as cache
type RedisLimiter struct {
	Client *redis.Client
}

// Limit is the implementation of Limiter interface
func (l RedisLimiter) Limit(key string, maxAttempts int, expiration time.Duration) error {
	hits, err := l.get(key)
	if err != nil {
		return err
	}

	if hits >= maxAttempts {
		ttl, err := l.ttl(key)
		if err != nil {
			return err
		}

		return MaxAttemptsError{ttl, hits}
	}

	hits, err = l.hit(key)

	if err != nil {
		return err
	}

	if hits >= maxAttempts {
		err = l.expire(key, expiration)

		if err != nil {
			return err
		}

		return MaxAttemptsError{expiration, hits}
	}

	return nil
}

func (l RedisLimiter) Reset(key string) error {
	cmd := l.Client.Del(key)
	return cmd.Err()
}

func (l RedisLimiter) get(key string) (hits int, err error) {
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

func (l RedisLimiter) hit(key string) (hits int, err error) {
	cmd := l.Client.Incr(key)
	err = cmd.Err()
	hits = int(cmd.Val())
	return
}

func (l RedisLimiter) expire(key string, duration time.Duration) error {
	cmd := l.Client.Expire(key, duration)
	return cmd.Err()
}

func (l RedisLimiter) ttl(key string) (time.Duration, error) {
	return l.Client.TTL(key).Result()
}
