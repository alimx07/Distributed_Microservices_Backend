package models

import "time"

type AppConfig struct {
	Server       ServerConfig             `yaml:"server"`
	RateLimiting RateLimitingConfig       `yaml:"rate_limiting"`
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
	RulesConfig string `yaml:"rules_config"`
	ScriptPath  string `yaml:"script_path"`
	Addr        string `yaml:"addr"`
	PoolSize    int    `yaml:"pool_size"`
}

type ServiceConfig struct {
	Instances           []string      `yaml:"instances"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	ProtosetPath        string        `yaml:"protoset_path"`
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
