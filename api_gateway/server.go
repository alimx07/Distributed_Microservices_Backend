package main

import (
	"net"
	"net/http"
)

type Server struct {
	config  Config
	router  *http.ServeMux
	handler *ApiHandler
}

func NewServer(h *ApiHandler, config Config) *Server {
	server := &Server{
		router:  http.NewServeMux(),
		handler: h,
	}
	server.addRoutes()
	return server
}

func (s *Server) start() error {
	var handler http.Handler = s.router
	handler = loggingMiddleware(handler)
	handler = authMiddleware(handler, s.config)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(s.config.ServerHost, s.config.ServerPort),
		Handler: handler,
	}

	return httpServer.ListenAndServe()
}

// I will use Static Routing for now
func (s *Server) addRoutes() {
	s.router.HandleFunc("/register", s.handler.RegisterHandler)
	s.router.HandleFunc("/login", s.handler.LoginHandler)
	s.router.HandleFunc("/profile/", s.handler.GetProfileHandler)

	// others later...
}
