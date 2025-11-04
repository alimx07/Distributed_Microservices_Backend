package main

import (
	"context"
	"encoding/json"
	"errors"
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

type redisPool struct {
	pool chan *redis.Client
}

type RateLimiter struct {
	ctx    context.Context
	rules  []Rule
	redis  *redisPool
	script string //lua script to run redis commands
}

type KeyExtractor func(r *http.Request) string

type Rule struct {
	Name       string `json:"name"`       // name of Rule (Ip , UserID , URL)
	Limit      int    `json:"limit"`      // bucket size
	RefillRate int    `json:"refillRate"` // requests/s
}

func NewRateLimiter(ctx context.Context, config models.RateLimitingConfig) (*RateLimiter, error) {
	p, err := newRedisPool(config.Addr, config.PoolSize)
	if err != nil {
		return nil, err
	}

	rules, err := loadRules(config.RulesConfig)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(config.ScriptPath)
	if err != nil {
		return nil, err
	}
	return &RateLimiter{ctx: ctx, rules: rules, redis: p, script: string(data)}, nil
}

// Create new Connection pool of size N
func newRedisPool(addr string, n int) (*redisPool, error) {
	var client *redis.Client
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	var numOfCon int
	rsPool := &redisPool{
		pool: make(chan *redis.Client, n),
	}
	for range n {
		client = redis.NewClient(&redis.Options{
			Addr: addr,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			continue
		}
		numOfCon++
		rsPool.pool <- client
	}
	if numOfCon == 0 {
		return nil, errors.New("no Connections opened with redis")
	}
	return rsPool, nil
}

func (rp *redisPool) Get() *redis.Client {
	return <-rp.pool
}

func (rp *redisPool) Put(r *redis.Client) {
	rp.pool <- r
}
func (rl *RateLimiter) Allow(r *http.Request) (bool, error) {
	var ok bool
	var err error

	keys, args := rl.matchRule(r)
	ok, err = rl.checkRedis(keys, args)

	return ok, err
}

// Check redis atomically using lua script
func (rl *RateLimiter) checkRedis(keys []string, args []interface{}) (bool, error) {

	// this could be blocking for ms
	// we could control that easily by tune pool size
	conn := rl.redis.Get()
	defer rl.redis.Put(conn)
	res, err := conn.Eval(rl.ctx, rl.script, keys, args).Result()

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
	return r.Header.Get("UserID")
}

func (rl *RateLimiter) matchRule(r *http.Request) ([]string, []interface{}) {
	// for now i have just two rules (Per IP , per UserID)
	// More custom rules (ex: per URL) can be added
	// just add different rules with name as the URL
	// and we can match them efficiently using data structure like trie

	// UserID header will be added by Auth if user is legal one

	var id string
	var keys []string
	var args []interface{}

	for _, rule := range rl.rules {
		if rule.Name == "IP" {
			id = ipExtractor(r)
			keys = []string{id}
			args = []interface{}{rule.RefillRate, rule.Limit, time.Now().Unix()}
		} else {
			id = userIdExtractor(r)
			keys = []string{id}
			args = []interface{}{rule.RefillRate, rule.Limit, time.Now().Unix()}
		}
	}

	// IN case of multiple Rules.
	// Id can be composite of userID + endpoint
	// like : User123:Post
	return keys, args
}

func loadRules(configPath string) ([]Rule, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var rules []Rule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
