package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/api_gateway/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

func LoadConfig() (models.ServerConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}
	config := models.ServerConfig{
		Host:          os.Getenv("SERVER_PORT"),
		Port:          os.Getenv("SERVER_HOST"),
		PublickeyAddr: os.Getenv("Public_Key_Addr"),
	}
	return config, nil
}
func InitLogger() {

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func loadLuaScripts(config *models.AppConfig) error {
	script, err := os.ReadFile(config.Redis.AddScriptPath)
	if err != nil {
		log.Println("Error in Loading ", config.Redis.AddScriptPath)
		return err
	}
	config.Redis.AddScript = string(script)
	script, err = os.ReadFile(config.Redis.CheckScriptPath)
	if err != nil {
		log.Println("Error in Loading ", config.Redis.CheckScript)
		return err
	}
	config.Redis.CheckScript = string(script)
	script, err = os.ReadFile(config.RateLimiting.ScriptPath)
	if err != nil {
		log.Println("Error in Loading ", config.RateLimiting.ScriptPath)
		return err
	}
	config.RateLimiting.RateLimitingScript = string(script)

	log.Println("---> All Lua Scripts Loaded successfully")
	return nil
}

type PublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
}

func GetPublicKey(addr string) ([]byte, error) {

	var err error

	resp, err := http.Get(addr)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(resp.Status)
		log.Println(resp.StatusCode)
		return nil, err
	}

	log.Println(resp.StatusCode)
	var data PublicKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println(err.Error())
		return nil, err
	}
	log.Println(data.PublicKey)
	return []byte(data.PublicKey), nil
}

func LoadAppConfig(filename string) (*models.AppConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config models.AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Override with env vars
	if serverHost := os.Getenv("SERVER_HOST"); serverHost != "" {
		config.Server.Host = serverHost
	}
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		config.Server.Port = serverPort
	}
	if publicKeyAddr := os.Getenv("PUBLIC_KEY_ADDR"); publicKeyAddr != "" {
		config.Server.PublickeyAddr = publicKeyAddr
	}
	if clusterAddr := os.Getenv("CLUSTER_ADDR"); clusterAddr != "" {
		log.Println(clusterAddr)
		clusterAddr := strings.Split(clusterAddr, ",")
		config.RateLimiting.Addr = clusterAddr
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		log.Println(redisAddr)
		config.Redis.RedisAddr = redisAddr
	}

	return &config, nil
}

// This func will validate token & make sure that it is not revoked
// by check redis instance
func ValidateToken(token string, pubKey []byte, r *redis.Client, luaScript string) (string, error) {
	if len(pubKey) == 0 {
		log.Println("Empty public key")
		return "", errors.New("empty PubKey")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tokenKey := "revoked:token:" + token
	result, err := r.Eval(ctx, luaScript, []string{tokenKey}).Result()
	if err != nil {
		log.Printf("Error checking token denylist: %v", err)
		// Fail-open again
	} else if res, ok := result.(int64); ok && res == 1 {
		log.Printf("Token Revoked: %v", token)
		return "", errors.New("invalid")
	}
	rsaPubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		log.Printf("Failed to parse RSA public key: %v", err)
		return "", errors.New("invalid public key")
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience("api_gateway"),
		jwt.WithIssuer("users_service"),
	)
	claims := jwt.RegisteredClaims{}
	parse, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return rsaPubKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Printf("Token expired: %v", token)
			return "", errors.New("invalid")
		}
		return "", err
	}
	if !parse.Valid {
		return "", errors.New("invalid")
	}
	if claims.Subject == "" {
		return "", errors.New("invalid")
	}
	return claims.Subject, nil
}

// type RedisPool struct {
// 	pool     chan *redis.Client
// 	numConns int
// }

// // Create new Connection pool of size N
// func NewRedisPool(addr string, n int) (*RedisPool, error) {
// 	log.Println("------>", addr, n)
// 	var client *redis.Client
// 	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// 	defer cancel()
// 	var numOfCon int
// 	rsPool := &RedisPool{
// 		pool: make(chan *redis.Client, n),
// 	}
// 	for range n {
// 		client = redis.NewClient(&redis.Options{
// 			Addr: addr,
// 		})
// 		if err := client.Ping(ctx).Err(); err != nil {
// 			continue
// 		}
// 		numOfCon++
// 		rsPool.pool <- client
// 	}
// 	if numOfCon == 0 {
// 		return nil, errors.New("no Connections opened with redis")
// 	}
// 	rsPool.numConns = numOfCon
// 	return rsPool, nil
// }

// func (rp *RedisPool) Get() *redis.Client {
// 	return <-rp.pool
// }

// func (rp *RedisPool) Put(r *redis.Client) {
// 	rp.pool <- r
// }

// func (rp *RedisPool) close() {
// 	var client *redis.Client
// 	for range rp.numConns {
// 		client = rp.Get()
// 		if err := client.Close(); err != nil {
// 			log.Printf("Error in Closing Client %v in redisPool", client.String())
// 		}
// 	}
// }
