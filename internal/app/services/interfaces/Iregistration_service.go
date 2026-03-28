package interfaces

import (
	"context"
	"system-discovery/internal/dto"
)

type IRegistrationService interface {
	Register(context context.Context, request dto.RegisterRequest)
	DeRegister(context context.Context, peer string)
}
