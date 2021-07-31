package api

import (
	"github.com/exelban/cheks/app/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pkgz/rest"
	"net/http"
	"time"
)

//go:generate moq -out mock_test.go . monitor

type monitor interface {
	Status() map[string]types.StatusType
	History() map[string]map[time.Time]bool
}

type Rest struct {
	Monitor monitor
}

func (s *Rest) Router() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.GetHead)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(120 * time.Second))

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300,
		AllowCredentials: true,
	})
	router.Use(corsOptions.Handler)

	router.Use(rest.Logger)
	router.NotFound(rest.NotFound)

	router.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		rest.JsonResponse(w, s.Monitor.Status())
	})
	router.Get("/history", func(w http.ResponseWriter, r *http.Request) {
		rest.JsonResponse(w, s.Monitor.History())
	})

	return router
}
