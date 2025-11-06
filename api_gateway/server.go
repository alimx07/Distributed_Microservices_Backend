package main

import (
	"log"
	"net"
	"net/http"

	"github.com/alimx07/Distributed_Microservices_Backend/api_gateway/models"
)

type Server struct {
	config  *models.AppConfig
	router  *http.ServeMux
	handler *Handler
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

	log.Printf("API Gateway starting on %s:%s", s.config.Server.Host, s.config.Server.Port)
	return httpServer.ListenAndServe()
}

// Add routes dynamically from configuration
func (s *Server) addRoutes() {
	// Register all routes from config
	for _, route := range s.config.Routes {
		pattern := route.Method + " " + route.Path
		s.router.HandleFunc(pattern, s.handler.GenericHandler)

		log.Printf("Registered route: %s %s -> %s.%s",
			route.Method, route.Path, route.GRPCService, route.GRPCMethod)
	}
	// Health check endpoint
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})
}
