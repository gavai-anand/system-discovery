package bootstrap

import (
	"context"
	"os"
	"system-discovery/internal/app/handlers"
	"system-discovery/internal/app/services"
)

type App struct {
	RegistrationHandler *handlers.RegistrationHandler
	DiscoveryHandler    *handlers.DiscoveryHandler
	HealthHandler       *handlers.HealthHandler
	CounterHandler      *handlers.CounterHandler

	ServiceCall         *services.ServiceCall
	DiscoveryService    *services.DiscoveryService
	RegistrationService *services.RegistrationService
	CounterService      *services.CounterService

	Self string
}

func NewApp(ctx context.Context) *App {
	self := os.Getenv("SELF")

	store := services.InitPeerStore()
	serviceCall := services.InitServiceCall()

	registrationService := services.InitRegistrationService(store)
	discoveryService := services.InitDiscoveryService(store, registrationService, self, serviceCall)
	counterService := services.InitCounterService(discoveryService, serviceCall, self)

	// Handlers
	registrationHandler := handlers.InitRegistrationHandler(registrationService)
	discoveryHandler := handlers.InitDiscoveryHandler(discoveryService)
	counterHandler := handlers.InitCounterHandler(counterService)
	healthHandler := handlers.InitHealthHandler()

	return &App{
		RegistrationHandler: registrationHandler,
		DiscoveryHandler:    discoveryHandler,
		HealthHandler:       healthHandler,
		CounterHandler:      counterHandler,

		ServiceCall:         serviceCall,
		DiscoveryService:    discoveryService,
		Self:                self,
		RegistrationService: registrationService,
		CounterService:      counterService,
	}
}
