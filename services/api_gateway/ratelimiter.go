package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/services/api_gateway/models"
	"github.com/redis/go-redis/v9"
)

// 1 - laod rules
// 2 - start get requests
// 3 - apply rules
// 4 - return nil or error

type RateLimiter struct {
	ctx          context.Context
	rules        map[string]Rule
	redisCluster *redis.ClusterClient
	script       string //lua script to run redis commands
}

type KeyExtractor func(r *http.Request) string

type Rule struct {
	Limit      int `json:"limit"`      // bucket size
	RefillRate int `json:"refillRate"` // requests/s
}

type RateLimitInfo struct {
	Allowed           bool
	Remaining         int
	Limit             int
	RetryAfterSeconds int
}

func NewRateLimiter(config models.RateLimitingConfig) (*RateLimiter, error) {
	c := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    config.Addr,
		PoolSize: config.RateLimiterPoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		log.Println("Error in Connection to redis Cluster: ", err)
		return nil, err
	}
	rules, err := loadRules(config.RulesConfig)
	if err != nil {
		return nil, err
	}
	ctx = context.Background()
	return &RateLimiter{ctx: ctx, rules: rules, redisCluster: c, script: config.RateLimitingScript}, nil
}

func (rl *RateLimiter) AllowIP(r *http.Request) (*RateLimitInfo, error) {
	id := ipExtractor(r)
	// log.Println("IP ID", id)
	return rl.Allow(id, rl.rules["IP"])
}

func (rl *RateLimiter) AllowRules(r *http.Request) (*RateLimitInfo, error) {
	userID := userIdExtractor(r)
	var mostRestrictive *RateLimitInfo
	for ruleName, rule := range rl.rules {
		// if ruleName == "IP" {
		// 	continue
		// }
		// for now i have just two rules (Per IP , per UserID)
		// More custom rules (ex: per URL) can be added
		// just add different rules with name as the URL
		// and we can match them efficiently using data structure like trie
		// matched := matchRule(r , rule.Name)
		info, err := rl.Allow(userID+ruleName, rule)
		if err != nil {
			return info, err
		}
		if !info.Allowed {
			return info, nil
		}
		// Keep track of most restrictive limits
		if mostRestrictive == nil || info.Remaining < mostRestrictive.Remaining {
			mostRestrictive = info
		}
	}
	if mostRestrictive == nil {
		return &RateLimitInfo{Allowed: true}, nil
	}
	return mostRestrictive, nil
}

func (rl *RateLimiter) Allow(id string, rule Rule) (*RateLimitInfo, error) {

	keys := []string{id}
	args := []interface{}{rule.RefillRate, rule.Limit, time.Now().Unix()}
	info, err := rl.checkRedis(keys, args)
	// log.Printf("REDIS RL: %v %v", info, err)
	if err != nil {
		return info, err
	}

	return info, nil
}

// Check redis atomically using lua script
func (rl *RateLimiter) checkRedis(keys []string, args []interface{}) (*RateLimitInfo, error) {

	res, err := rl.redisCluster.Eval(rl.ctx, rl.script, keys, args).Result()

	log.Println("REDIS RESULT", res)

	// Now it is a trade off
	// lets fail-Open
	if err != nil {
		log.Printf("There is error in redis connection: %v", err.Error())
		return &RateLimitInfo{Allowed: true}, err
	}

	// Parse result array: [allowed, remaining, limit, retry_after]
	if resultArray, ok := res.([]interface{}); ok && len(resultArray) >= 4 {
		allowed := false
		if allowedVal, ok := resultArray[0].(int64); ok {
			allowed = allowedVal == 1
		}

		remaining := 0
		if remainingVal, ok := resultArray[1].(int64); ok {
			remaining = int(remainingVal)
		}

		limit := 0
		if limitVal, ok := resultArray[2].(int64); ok {
			limit = int(limitVal)
		}

		retryAfter := 0
		if retryAfterVal, ok := resultArray[3].(int64); ok {
			retryAfter = int(retryAfterVal)
		}

		return &RateLimitInfo{
			Allowed:           allowed,
			Remaining:         remaining,
			Limit:             limit,
			RetryAfterSeconds: retryAfter,
		}, nil
	}

	// Fallback for old format or unexpected result
	if v, ok := res.(int64); ok {
		return &RateLimitInfo{
			Allowed:   v == 1,
			Remaining: 0,
			Limit:     0,
		}, nil
	}
	return &RateLimitInfo{Allowed: true}, nil
}

func ipExtractor(r *http.Request) string {
	return r.RemoteAddr
}

func userIdExtractor(r *http.Request) string {
	return r.Header.Get("UserId")
}

func loadRules(configPath string) (map[string]Rule, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var rules map[string]Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *RateLimiter) close() {
	if err := r.redisCluster.Close(); err != nil {
		log.Println("Closing rateLimiter Error: ", err.Error())
	}
}
