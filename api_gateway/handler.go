package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alimx07/Distributed_Microservices_Backend/api_gateway/models"
)

type Handler struct {
	config        *models.AppConfig
	loadBalancers map[string]*RoundRobin
	grpcInvoker   *GRPCInvoker
	rateLimiter   *RateLimiter
	redis         *RedisPool
	routeMap      map[string]map[string]*models.RouteConfig // method -> path -> config
}

func NewHandler(config *models.AppConfig, loadBalancers map[string]*RoundRobin, grpcInvoker *GRPCInvoker, rateLimiter *RateLimiter, redis *RedisPool) *Handler {
	h := &Handler{
		config:        config,
		loadBalancers: loadBalancers,
		grpcInvoker:   grpcInvoker,
		rateLimiter:   rateLimiter,
		redis:         redis,
		routeMap:      make(map[string]map[string]*models.RouteConfig),
	}
	var err error

	log.Println(config.Server.PublickeyAddr)
	config.PublicKey, err = GetPublicKey(config.Server.PublickeyAddr)

	if err != nil {
		log.Println("Error in Loading the public key")
		return nil
	}
	// Build route map for fast lookup
	for i := range config.Routes {
		route := &config.Routes[i]
		if h.routeMap[route.Method] == nil {
			h.routeMap[route.Method] = make(map[string]*models.RouteConfig)
		}
		h.routeMap[route.Method][route.Path] = route
	}

	return h
}

// GenericHandler handles all HTTP requests and routes them to the required gRPC services
func (h *Handler) GenericHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s is requested\n", r.URL.Path)

	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "localhost") // mock url for now
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Find matching route
	route := h.findRoute(r.Method, r.URL.Path)
	if route == nil {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	// Apply rate limiting if enabled
	if route.RateLimitEnabled {
		allowed, err := h.rateLimiter.Allow(r)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			// Fail open
		} else if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	// Check authentication if required
	var userID string
	if route.RequireAuth {
		var ok bool
		userID, ok = h.checkAuth(w, r)
		if !ok {
			return // Auth middleware already wrote error
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Build Request body , add any params found
	requestData, err := h.buildRequestData(r, route, body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build request: %v", err), http.StatusBadRequest)
		return
	}

	lb, exists := h.loadBalancers[route.Service]
	if !exists {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	// Get healthy connection from load balancer
	conn, err := lb.ServiceConn()
	if err != nil {
		log.Printf("Failed to get service connection: %v", err)
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	// Invoke gRPC method dynamically
	ctx := context.Background()
	responseJSON, err := h.grpcInvoker.Invoke(
		ctx,
		conn,
		route.GRPCService,
		route.GRPCMethod,
		requestData,
		userID,
	)
	if err != nil {
		log.Printf("gRPC invocation error: %v", err)
		http.Error(w, fmt.Sprintf("Service error: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO:
	// Try to refactor and find more modular way to do that
	if strings.HasSuffix(route.Path, "login") || strings.HasSuffix(route.Path, "refresh") {
		var token models.Tokens

		if err := json.Unmarshal(responseJSON, &token); err != nil {
			log.Println("tokens unmarshal failed")
			http.Error(w, fmt.Sprintf("Decoding Error , %v", err), http.StatusInternalServerError)
		}

		access_token := &http.Cookie{
			Name:     "access_token",
			Value:    token.Access,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			//  Secure:   true,  required in production for https
		}

		refresh_token := &http.Cookie{
			Name:  "refresh_token",
			Value: token.Refresh,
			// TODO:
			// I can send this token only into some path like api/v1/refresh
			// Is there a way to send to only two paths instead of duplicate the token or not ?
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			//  Secure:   true,  required in production for https
		}
		http.SetCookie(w, access_token)
		http.SetCookie(w, refresh_token)
	}

	if strings.HasSuffix(route.Path, "logout") {
		var token models.Tokens

		if err := json.Unmarshal(requestData, &token); err != nil {
			log.Println("tokens unmarshal failed")
			http.Error(w, fmt.Sprintf("Decoding Error , %v", err), http.StatusInternalServerError)
		}
		rc := h.redis.Get()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rc.Eval(ctx, h.config.Redis.AddScript, []string{token.Access}, []interface{}{5 * time.Minute})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// findRoute finds the matching route configuration
func (h *Handler) findRoute(method, path string) *models.RouteConfig {

	// In Case of wrong Method required
	_, exists := h.routeMap[method]
	if !exists {
		return nil
	}

	// First try exact match
	// If found
	if route, exists := h.routeMap[method][path]; exists {
		return route
	}

	// Try pattern matching for paths with parameters
	for routePath, route := range h.routeMap[method] {
		if h.matchPath(routePath, path) {
			return route
		}
	}

	return nil
}

// matchPath checks if a request path matches a route pattern
// TODO:
// --> Use Trie for faster match
func (h *Handler) matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	// like in case
	// api/v1/register != api/v1/followers/{userId}
	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}") {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}

	return true
}

// buildRequestData builds the gRPC request from HTTP request
func (h *Handler) buildRequestData(r *http.Request, route *models.RouteConfig, body []byte) ([]byte, error) {
	var data map[string]interface{}

	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			data = make(map[string]interface{})
		}
	} else {
		data = make(map[string]interface{})
	}
	// extract header/Tokens
	tokenParams := h.extractTokens(r)

	for key, value := range tokenParams {
		data[key] = value
	}
	// For patterns like /api/v1/posts/{postId}, this extracts "postId"
	pathParams := h.extractPathParams(route.Path, r)
	for key, value := range pathParams {
		data[key] = value
	}

	// query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			data[key] = values[0]
		}
	}

	return json.Marshal(data)
}

func (h *Handler) extractTokens(r *http.Request) map[string]interface{} {
	tokens := make(map[string]interface{})
	token := r.Header.Get("refresh_token")
	if token == "" {
		return tokens
	}
	tokens["refresh_token"] = token
	return tokens
}

func (h *Handler) extractPathParams(pattern string, r *http.Request) map[string]interface{} {
	params := make(map[string]interface{})

	patternParts := strings.Split(pattern, "/")

	for _, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			if value := r.PathValue(paramName); value != "" {
				params[paramName] = value
			}
		}
	}

	return params
}

func (h *Handler) checkAuth(w http.ResponseWriter, r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return "", false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	if len(token) == 0 {
		http.Error(w, "Authorization token required", http.StatusUnauthorized)
		return "", false
	}

	userID, err := ValidateToken(token, h.config.PublicKey, h.redis.Get(), h.config.Redis.CheckScript)
	if err != nil {
		if err.Error() == "invalid" {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		return "", false
	}

	return userID, true
}
