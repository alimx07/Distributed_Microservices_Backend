package models

import "time"

type AppConfig struct {
	Server       ServerConfig             `yaml:"server"`
	RateLimiting RateLimitingConfig       `yaml:"rate_limiting"`
	Redis        RedisConfig              `yaml:"redis_config"`
	Services     map[string]ServiceConfig `yaml:"services"`
	Routes       []RouteConfig            `yaml:"routes"`
	PublicKey    []byte
}

type ServerConfig struct {
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	PublickeyAddr string `yaml:"public_key_addr"`
}

type RateLimitingConfig struct {
	RulesConfig         string `yaml:"rules_config"`
	ScriptPath          string `yaml:"script_path"`
	Addr                string `yaml:"addr"`
	RateLimiterPoolSize int    `yaml:"pool_size"`
	RateLimitingScript  string
}

type ServiceConfig struct {
	Instances           []string      `yaml:"instances"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	ProtosetPath        string        `yaml:"protoset_path"`
}

type RedisConfig struct {
	RedisAddr       string `yaml:"redis_addr"`
	AddScriptPath   string `yaml:"redis_add_script"`
	CheckScriptPath string `yaml:"redis_check_script"`
	RedisPoolSize   int    `yaml:"redis_pool_size"`
	AddScript       string
	CheckScript     string
}

type RouteConfig struct {
	Path             string `yaml:"path"`
	Method           string `yaml:"method"`
	Service          string `yaml:"service"`
	GRPCService      string `yaml:"grpc_service"`
	GRPCMethod       string `yaml:"grpc_method"`
	RequireAuth      bool   `yaml:"require_auth"`
	RateLimitEnabled bool   `yaml:"rate_limit_enabled"`
}

type User struct {
	Username string
	Email    string
	Password string
}

type Tokens struct {
	Access  string
	Refresh string
}
