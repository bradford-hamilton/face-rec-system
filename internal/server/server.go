package server

import (
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.mlbam.net/blamson/face-rec-system/internal/storage"
)

// API is a structure that orchestrates the http layer and database.
type API struct {
	baseURL string
	db      storage.Database
	Mux     *chi.Mux
}

// New takes a storage.Repository and set's up an API server, using the store.
func New(db storage.Database) *API {
	r := chi.NewRouter()
	r.Use(
		corsMiddleware().Handler,
		middleware.Logger,
		middleware.StripSlashes,            // strip slashes to no slash URL versions
		middleware.Recoverer,               // recover from panics without crashing server
		middleware.Timeout(30*time.Second), // start with a pretty standard timeout
	)

	baseURL := "http://localhost:4000"
	if os.Getenv("FACE_REC_SYSTEM_ENVIRONMENT") == "production" {
		baseURL = "NULL"
	}

	api := &API{db: db, Mux: r, baseURL: baseURL}
	api.initializeRoutes()

	return api
}

func (a *API) initializeRoutes() {
	a.Mux.Post("/register", a.RegisterHandler)
	a.Mux.Post("/match", a.MatchHandler)
}

func corsMiddleware() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		MaxAge:         300, // Maximum value not ignored by any major browsers
	})
}
