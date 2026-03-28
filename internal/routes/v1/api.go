package v1

import (
	"net/http"
	"system-discovery/internal/bootstrap"
	"system-discovery/internal/middleware"
)

func Router(app *bootstrap.App) http.Handler {
	//Its responsible for the matching URL path with corresponding handler functions
	mux := http.NewServeMux()

	//Base path
	base := "/api/v1"

	mux.HandleFunc(base+"/register", MethodHandler(http.MethodPost, app.RegistrationHandler.RegisterNode))
	mux.HandleFunc(base+"/peer-list", MethodHandler(http.MethodGet, app.DiscoveryHandler.GetAllPeers))
	mux.HandleFunc(base+"/sync-peers", MethodHandler(http.MethodPost, app.DiscoveryHandler.SyncPeers))
	mux.HandleFunc(base+"/health", MethodHandler(http.MethodGet, app.HealthHandler.Health))
	mux.HandleFunc(base+"/increment", MethodHandler(http.MethodPost, app.CounterHandler.Increment))
	mux.HandleFunc(base+"/replicate", MethodHandler(http.MethodPost, app.CounterHandler.Replicate))
	mux.HandleFunc(base+"/count", MethodHandler(http.MethodGet, app.CounterHandler.GetCount))
	mux.HandleFunc(base+"/operations", MethodHandler(http.MethodGet, app.CounterHandler.GetOperations))

	return middleware.LoggingMiddleware(mux)
}

func MethodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}
