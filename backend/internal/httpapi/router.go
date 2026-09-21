package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/bescobarm/duo-taste/backend/internal/config"
	"github.com/bescobarm/duo-taste/backend/internal/store"
)

// NewRouter wires every route and the middleware chain around them.
func NewRouter(places store.PlaceStore, cfg config.Config, logger *slog.Logger) http.Handler {
	handler := placeHandler{places: places}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /api/categories", listCategories)
	mux.HandleFunc("GET /api/places", handler.list)
	mux.HandleFunc("POST /api/places", handler.create)
	mux.HandleFunc("GET /api/places/{id}", handler.get)
	mux.HandleFunc("POST /api/places/{id}/ratings", handler.addRating)

	return requestLog(logger, cors(cfg.AllowOrigin, mux))
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
