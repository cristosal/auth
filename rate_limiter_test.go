package auth

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

var url = "localhost:5973"

func TestRateLimiter(t *testing.T) {
	rds := redis.NewClient(&redis.Options{
		Addr:       url,
		MaxRetries: 3,
	})

	limiter := RedisLimiter{Client: rds}
	limiter.Limit("test_key", 1, time.Duration(time.Second))
}
