package handlers

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"system-discovery/internal/app/services/interfaces"
	"system-discovery/internal/dto"
	"system-discovery/internal/utils"
)

// RegistrationHandler handles node registration in the cluster.
type RegistrationHandler struct {
	BaseHandler
	rs interfaces.IRegistrationService
}

func InitRegistrationHandler(rs interfaces.IRegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		rs: rs,
	}
}

// RegisterNode registers a new node after validating the request payload.
func (rh *RegistrationHandler) RegisterNode(w http.ResponseWriter, r *http.Request) {
	var request dto.RegisterRequest

	// Decode request body
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("[ERROR] Invalid JSON in register request: %v", err)
		rh.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("Invalid Json: %s", err.Error()))
		return
	}

	// Validate request fields
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		log.Printf("[ERROR] Validation failed for register request: %v", err)
		rh.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("Invalid Request: %s", err.Error()))
		return
	}

	ctx := r.Context()

	// Register node in the system
	rh.rs.Register(ctx, request)

	rh.ResponseOK(w, nil)
}
