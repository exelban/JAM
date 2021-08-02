package api

import (
	"github.com/exelban/cheks/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pkgz/rest"
	"html/template"
	"log"
	"net/http"
	"time"
)

//go:generate moq -out mock_test.go . monitor

type monitor interface {
	Status() map[string]types.StatusType
	Services() map[string]types.Service
}

type Rest struct {
	Monitor  monitor
	Version  string
	Live     bool
	Template *template.Template
}

var indexPath = "index.html"

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

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmp := s.Template

		if s.Live {
			t, err := template.ParseFiles(indexPath)
			if err != nil {
				log.Printf("[ERROR] parse html %v", err)
				rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, err.Error())
				return
			}
			tmp = t
		}

		items := struct {
			Version string
			List    map[string]types.Service
		}{
			Version: s.Version,
			List:    s.Monitor.Services(),
		}

		if err := tmp.Execute(w, items); err != nil {
			log.Printf("[ERROR] render html %v", err)
			rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, err.Error())
		}
	})

	router.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		rest.JsonResponse(w, s.Monitor.Status())
	})

	return router
}
