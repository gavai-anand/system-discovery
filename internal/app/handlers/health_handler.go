package handlers

import (
	"net/http"
)

type HealthHandler struct {
	BaseHandler
}

func InitHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (hh HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	hh.ResponseOK(w, nil)
}
