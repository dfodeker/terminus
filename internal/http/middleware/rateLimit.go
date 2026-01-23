// internal/http/middleware/ratelimit.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis  *redis.Client
	limits map[string]RateLimit // Plan -> Limit
}

type RateLimit struct {
	RequestsPerSecond int
	BurstSize         int
}

var DefaultLimits = map[string]RateLimit{
	"free":       {RequestsPerSecond: 2, BurstSize: 10},
	"starter":    {RequestsPerSecond: 4, BurstSize: 20},
	"pro":        {RequestsPerSecond: 10, BurstSize: 50},
	"enterprise": {RequestsPerSecond: 100, BurstSize: 500},
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		limits: DefaultLimits,
	}
}

func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := AuthFromContext(ctx)

		if auth == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get rate limit for shop's plan
		limit, ok := rl.limits[auth.Shop.PlanName]
		if !ok {
			limit = rl.limits["free"]
		}

		// Sliding window rate limit using Redis
		key := fmt.Sprintf("ratelimit:%s", auth.ShopID)

		allowed, remaining, resetAt, err := rl.checkLimit(ctx, key, limit)
		if err != nil {
			// On Redis error, allow request but log
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit.BurstSize))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(resetAt-time.Now().Unix(), 10))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) checkLimit(ctx context.Context, key string, limit RateLimit) (allowed bool, remaining int, resetAt int64, err error) {
	now := time.Now()
	windowStart := now.Add(-time.Second).UnixMicro()

	// Lua script for atomic sliding window
	script := redis.NewScript(`
        local key = KEYS[1]
        local now = tonumber(ARGV[1])
        local window = tonumber(ARGV[2])
        local limit = tonumber(ARGV[3])
        
        -- Remove old entries
        redis.call('ZREMRANGEBYSCORE', key, '-inf', window)
        
        -- Count current entries
        local count = redis.call('ZCARD', key)
        
        if count < limit then
            -- Add new entry
            redis.call('ZADD', key, now, now .. '-' .. math.random())
            redis.call('EXPIRE', key, 2)
            return {1, limit - count - 1}
        else
            return {0, 0}
        end
    `)

	result, err := script.Run(ctx, rl.redis, []string{key},
		now.UnixMicro(), windowStart, limit.BurstSize).Slice()
	if err != nil {
		return false, 0, 0, err
	}

	allowed = result[0].(int64) == 1
	remaining = int(result[1].(int64))
	resetAt = now.Add(time.Second).Unix()

	return allowed, remaining, resetAt, nil
}
