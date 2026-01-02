package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/api_gateway/models"
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

func (rl *RateLimiter) AllowIP(r *http.Request) (bool, error) {
	id := ipExtractor(r)
	// log.Println("IP ID", id)
	return rl.Allow(id, rl.rules["IP"])
}

func (rl *RateLimiter) AllowRules(r *http.Request) (bool, error) {
	userID := userIdExtractor(r)
	for ruleName, rule := range rl.rules {
		if ruleName == "IP" {
			continue
		}
		// for now i have just two rules (Per IP , per UserID)
		// More custom rules (ex: per URL) can be added
		// just add different rules with name as the URL
		// and we can match them efficiently using data structure like trie
		// matched := matchRule(r , rule.Name)
		ok, err := rl.Allow(userID+ruleName, rule)
		if err != nil && !ok {
			return ok, err
		}
	}
	return true, nil
}

func (rl *RateLimiter) Allow(id string, rule Rule) (bool, error) {

	keys := []string{id}
	args := []interface{}{rule.RefillRate, rule.Limit, time.Now().Unix()}
	ok, err := rl.checkRedis(keys, args)
	// log.Printf("REDIS RL: %v %v", ok, err)
	if err != nil && !ok {
		return false, err
	}

	return true, nil
}

// Check redis atomically using lua script
func (rl *RateLimiter) checkRedis(keys []string, args []interface{}) (bool, error) {

	res, err := rl.redisCluster.Eval(rl.ctx, rl.script, keys, args).Result()

	log.Println("REDIS RESULT", res)

	// Now it is a trade off
	// lets fail-Open
	if err != nil {
		log.Printf("There is error in redis connection: %v", err.Error())
		return true, err
	}
	if v, ok := res.(int64); ok {
		return v == 1, nil
	}
	return true, nil
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
