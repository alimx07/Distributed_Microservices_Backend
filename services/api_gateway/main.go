package main

import (
	"context"
	"log"
	"net/http"
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

	LoadBalancer, err := NewLoadBalancer(config.ServiceRegistery)
	if err != nil {
		rateLimiter.close()
		log.Fatalf("Failed to initiallize LoadBalancer: %v", err)
	}
	// for serviceName, serviceConfig := range config.Services {
	// 	if len(serviceConfig.Instances) == 0 {
	// 		log.Printf("Warning: No instances configured for service %s", serviceName)
	// 		continue
	// 	}

	// 	lb := NewRoundRobin(serviceConfig.Instances, serviceConfig.HealthCheckInterval)
	// 	loadBalancers[serviceName] = lb
	// 	log.Printf("Load balancer initialized for %s with %d instances", serviceName, len(serviceConfig.Instances))
	// }

	grpcInvoker := NewGRPCInvoker(config.RouteOptions)
	log.Println("gRPC invoker initialized")

	for serviceName, protofile := range config.ProtoFiles {
		if protofile == "" {
			log.Printf("Warning: No protoset path configured for service %s", serviceName)
			continue
		}

		err := grpcInvoker.LoadProtoset(protofile, serviceName)
		if err != nil {
			log.Printf("Warning: Failed to load protoset for service %s: %v", serviceName, err)
		} else {
			log.Printf("Successfully loaded protoset for %s from %s", serviceName, protofile)
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
	handler := NewHandler(config, LoadBalancer, grpcInvoker, rateLimiter, redis)
	if handler == nil {
		rateLimiter.close()
		// grpcInvoker.close()
		LoadBalancer.close()
		redis.Close()
		log.Fatal("Failed to create handler")
	}

	// Initialize and start server
	server := NewServer(handler, config)
	log.Printf("Starting API Gateway on %s:%s", config.Server.Host, config.Server.Port)

	errChan := make(chan error, 1)
	go func() {
		if err := server.start(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("ShutDown Signal received")
	case err := <-errChan:
		log.Println("Server Error: ", err.Error())
	}
	stop()

	server.Close()
}
