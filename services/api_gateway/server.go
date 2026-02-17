package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/services/api_gateway/models"
)

type Server struct {
	config     *models.AppConfig
	router     *http.ServeMux
	httpServer *http.Server
	handler    *Handler
	serviceOFF atomic.Bool
}

func NewServer(handler *Handler, config *models.AppConfig) *Server {
	server := &Server{
		router:  http.NewServeMux(),
		handler: handler,
		config:  config,
	}
	server.addRoutes()
	return server
}

func (s *Server) start() error {
	var handler http.Handler = s.router

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.config.Server.Host, s.config.Server.Port),
		Handler: handler,
	}

	s.serviceOFF.Store(false)
	log.Printf("API Gateway starting on %s:%s", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = httpServer
	return httpServer.ListenAndServe()
}

func (s *Server) addRoutes() {
	// Get routes from handler's route map (built from google.api.http annotations)
	routeMap := s.handler.GetRouteMap()
	for method, routes := range routeMap {
		for path, route := range routes {
			pattern := method + " " + path
			s.router.HandleFunc(pattern, s.handler.GenericHandler)

			log.Printf("Registered route: %s %s -> %s.%s",
				method, path, route.GRPCService, route.GRPCMethod)
		}
	}
	// Health check endpoint
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if s.serviceOFF.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "unhealth"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "healthy"}`))
		}
	})
}

func (s *Server) Close() {
	// prevent server from get newConns
	s.serviceOFF.Store(true)

	// wait untill reflected in front LoadBalancer healthChecker
	time.Sleep(5 * time.Second)

	// Make sure current requests served

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// This also will ensure that all GrpcRequests Ended successfully
	// as they are part of GenericHandler in all httpRequests
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Println("Error in Closing HttpServer gracefully: ", err.Error())
		}
		log.Println("HttpServer Closed Successfully")
	}

	// Close any open resources that controlled by handler
	s.handler.close()

	// Closed Finally
}
