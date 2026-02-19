package models

import "time"

type AppConfig struct {
	Server       ServerConfig       `yaml:"server"`
	RateLimiting RateLimitingConfig `yaml:"rate_limiting"`
	Redis        RedisConfig        `yaml:"redis_config"`
	// ServiceRegistery RegisteryConfig         `yaml:"service_registery"`
	K8sServices  map[string]string       `yaml:"k8s_services"`
	ProtoFiles   map[string]string       `yaml:"protoset_files"`
	RouteOptions map[string]*RouteOption `yaml:"route_options"`
	PublicKey    []byte
}

type ServerConfig struct {
	Host          string `yaml:"host"`
	Port          string `yaml:"port"`
	PublickeyAddr string `yaml:"public_key_addr"`
}

type RateLimitingConfig struct {
	RulesConfig         string   `yaml:"rules_config"`
	ScriptPath          string   `yaml:"script_path"`
	Addr                []string `yaml:"addrs"`
	RateLimiterPoolSize int      `yaml:"pool_size"`
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
	AddScript       string
	CheckScript     string
}

type RegisteryConfig struct {
	ServiceRegisteryPath   string `yaml:"service_registery_path"`
	ServiceRegisteryPrefix string `yaml:"service_registery_prefix"`
}

type RouteOption struct {
	RequireAuth      bool `yaml:"require_auth"`
	RateLimitEnabled bool `yaml:"rate_limit_enabled"`
}

type RouteConfig struct {
	Path             string
	Method           string
	Body             string
	GRPCService      string
	GRPCMethod       string
	BackendService   string
	RequireAuth      bool
	RateLimitEnabled bool
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
