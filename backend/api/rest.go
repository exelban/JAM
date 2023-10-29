package api

import (
	"context"
	"github.com/exelban/uptime/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pkgz/rest"
	"net/http"
	"time"
)

type monitor interface {
	Services() []types.Service
}

type Rest struct {
	Monitor monitor
}

func (s *Rest) Router(ctx context.Context) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.GetHead)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(120 * time.Second))

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	router.Use(corsOptions.Handler)

	router.Use(rest.Logger)
	router.NotFound(rest.NotFound)

	router.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		rest.JsonResponse(w, s.Monitor.Services())
	})

	return router
}
