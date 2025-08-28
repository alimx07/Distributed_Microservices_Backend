package main

import (
	"context"
	"log"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler, config Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/register" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
		}
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			http.Error(w, "Authorization Header Required", http.StatusUnauthorized)
			return
		}
		userID, err := ValidateToken(authToken, config)
		if err != nil {
			http.Error(w, "Invalid/Expired Authorization Token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "UserID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
