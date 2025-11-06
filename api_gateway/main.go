package main

import (
	"context"
	"log"
)

func main() {
	// Initialize logger
	InitLogger()

	ctx := context.Background()

	config, err := LoadAppConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	rateLimiter, err := NewRateLimiter(ctx, config.RateLimiting)
	if err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}
	log.Println("Rate limiter initialized")

	loadBalancers := make(map[string]*RoundRobin)
	for serviceName, serviceConfig := range config.Services {
		if len(serviceConfig.Instances) == 0 {
			log.Printf("Warning: No instances configured for service %s", serviceName)
			continue
		}

		lb := NewRoundRobin(ctx, serviceConfig.Instances, serviceConfig.HealthCheckInterval)
		loadBalancers[serviceName] = lb
		log.Printf("Load balancer initialized for %s with %d instances", serviceName, len(serviceConfig.Instances))
	}

	grpcInvoker := NewGRPCInvoker()
	log.Println("gRPC invoker initialized")

	redis, err := NewRedisPool(config.Redis.RedisAddr, config.Redis.PoolSize)

	if err != nil {
		// Do not kill the service. just log
		log.Println("Warning: Redis Connection failed for DenyList instance")
	}

	for serviceName, serviceConfig := range config.Services {
		if serviceConfig.ProtosetPath == "" {
			log.Printf("Warning: No protoset path configured for service %s", serviceName)
			continue
		}

		err := grpcInvoker.LoadProtoset(serviceConfig.ProtosetPath)
		if err != nil {
			log.Printf("Warning: Failed to load protoset for service %s: %v", serviceName, err)
		} else {
			log.Printf("Successfully loaded protoset for %s from %s", serviceName, serviceConfig.ProtosetPath)
		}
	}

	handler := NewHandler(config, loadBalancers, grpcInvoker, rateLimiter, redis)
	log.Println("Handler initialized")

	// Initialize and start server
	server := NewServer(handler, config)
	log.Printf("Starting API Gateway on %s:%s", config.Server.Host, config.Server.Port)

	if err := server.start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
