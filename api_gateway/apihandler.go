package main

import (
	"net/http"
)

type ApiHandler struct {
}

func NewApiHandler() *ApiHandler {
	return &ApiHandler{}
}

func (h *ApiHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *ApiHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *ApiHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {

}
