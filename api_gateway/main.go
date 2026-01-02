package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// Initialize logger
	InitLogger()

	// ctx := context.Background()

	config, err := LoadAppConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	err = loadLuaScripts(config)
	if err != nil {
		log.Fatalf("Failed to LuaScripts : %v", err)
	}
	rateLimiter, err := NewRateLimiter(config.RateLimiting)
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

		lb := NewRoundRobin(serviceConfig.Instances, serviceConfig.HealthCheckInterval)
		loadBalancers[serviceName] = lb
		log.Printf("Load balancer initialized for %s with %d instances", serviceName, len(serviceConfig.Instances))
	}

	grpcInvoker := NewGRPCInvoker()
	log.Println("gRPC invoker initialized")

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

	// redisPool, err := NewRedisPool(config.Redis.RedisAddr, config.Redis.RedisPoolSize)
	// if err != nil {
	// 	rateLimiter.close()
	// 	for _, lb := range loadBalancers {
	// 		lb.Close()
	// 	}
	// 	log.Fatalf("Failed to create Redis pool: %v", err)
	// }
	redis := redis.NewClient(&redis.Options{Addr: config.Redis.RedisAddr})
	handler := NewHandler(config, loadBalancers, grpcInvoker, rateLimiter, redis)
	if handler == nil {
		rateLimiter.close()
		// grpcInvoker.close()
		for _, lb := range loadBalancers {
			lb.Close()
		}
		redis.Close()
		log.Fatal("Failed to create handler")
	}

	// Initialize and start server
	server := NewServer(handler, config)
	log.Printf("Starting API Gateway on %s:%s", config.Server.Host, config.Server.Port)

	go func() {
		log.Println(server.start())
	}()

	<-ctx.Done()

	stop()

	server.Close()
}
