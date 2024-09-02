package api

import (
	"context"
	"crypto/subtle"
	"fmt"
	"github.com/exelban/uptime/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lithammer/shortuuid/v4"
	"github.com/pkgz/rest"
	"io/fs"
	"log"
	"net/http"
	"time"
)

type monitor interface {
	Services() []types.Service
}

type config interface {
	FindHost(context.Context, string) (*types.Host, error)
	HostsList(context.Context) ([]*types.Host, error)

	AddHost(context.Context, *types.Host) error
	UpdateHost(context.Context, *types.Host) error
	DeleteHost(context.Context, string) error
}

type Authorization struct {
	Enabled bool

	Username string
	Password string
}

type Rest struct {
	FS fs.FS

	Monitor monitor
	Config  config
	Auth    Authorization
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
	router.Use(s.basicAuth)

	router.HandleFunc("/ui", s.ui)
	router.HandleFunc("/ui/*", s.ui)

	router.Route("/target", func(r chi.Router) {
		r.Get("/", s.targets)
		r.Post("/", s.addTarget)
		r.Post("/{id}", s.editTarget)
		r.Delete("/{id}", s.deleteTarget)
	})

	return router
}

func (s *Rest) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.Auth.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok || username != s.Auth.Username || subtle.ConstantTimeCompare([]byte(password), []byte(s.Auth.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			rest.ErrorResponse(w, r, http.StatusUnauthorized, nil, "restricted")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Rest) ui(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(s.FS)).ServeHTTP(w, r)
}

func (s *Rest) targets(w http.ResponseWriter, r *http.Request) {
	monitors := s.Monitor.Services()
	hosts, err := s.Config.HostsList(r.Context())
	if err != nil {
		log.Printf("[ERROR] hosts list: %v", err)
		rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "hosts list")
		return
	}

	type target struct {
		ID string `json:"id"`

		types.Service
		*types.Host
	}
	list := make([]target, len(monitors))

	for i, service := range monitors {
		list[i].ID = service.ID
		list[i].Service = service
		for _, host := range hosts {
			if service.ID == host.ID {
				list[i].Host = host
				break
			}
		}
	}

	rest.JsonResponse(w, list)
}

func (s *Rest) addTarget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type types.HostType `json:"type" yaml:"-"`

		Name string   `json:"name"`
		Tags []string `json:"tags"`

		Method string `json:"method"`
		URL    string `json:"url"`

		Retry            int `json:"retry"`
		Timeout          int `json:"timeout"`
		InitialDelay     int `json:"initialDelay"`
		SuccessThreshold int `json:"successThreshold"`
		FailureThreshold int `json:"failureThreshold"`
	}
	if err := rest.ReadBody(r, &req); err != nil {
		rest.ErrorResponse(w, r, http.StatusBadRequest, err, "unmarshal")
		return
	}

	h := &types.Host{
		ID:   shortuuid.New(),
		Type: req.Type,

		Name: req.Name,
		Tags: req.Tags,

		Method: req.Method,
		URL:    req.URL,

		SuccessThreshold: req.SuccessThreshold,
		FailureThreshold: req.FailureThreshold,
	}

	if req.Retry != 0 {
		h.Retry = fmt.Sprintf("%ds", req.Retry)
	}
	if req.Timeout != 0 {
		h.Timeout = fmt.Sprintf("%ds", req.Timeout)
	}
	if req.InitialDelay != 0 {
		h.InitialDelay = fmt.Sprintf("%ds", req.InitialDelay)
	}

	if err := s.Config.AddHost(r.Context(), h); err != nil {
		log.Printf("[ERROR] add host: %v", err)
		rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "add host")
		return
	}

	rest.OkResponse(w)
}

func (s *Rest) editTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	host, err := s.Config.FindHost(ctx, id)
	if err != nil {
		log.Printf("[ERROR] find host: %v", err)
		rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "find host")
		return
	}

	var req struct {
		Type *types.HostType `json:"type" yaml:"-"`

		Name *string `json:"name"`

		Method *string `json:"method"`
		URL    *string `json:"url"`
	}
	if err := rest.ReadBody(r, &req); err != nil {
		rest.ErrorResponse(w, r, http.StatusBadRequest, err, "unmarshal")
		return
	}

	updated := false
	if req.Type != nil && *req.Type != host.Type {
		host.Type = *req.Type
		updated = true
	}
	if req.Name != nil && *req.Name != host.Name {
		host.Name = *req.Name
		updated = true
	}
	if req.Method != nil && *req.Method != host.Method {
		host.Method = *req.Method
		updated = true
	}
	if req.URL != nil && *req.URL != host.URL {
		host.URL = *req.URL
		updated = true
	}

	if !updated {
		rest.OkResponse(w)
		return
	}

	if err := s.Config.UpdateHost(ctx, host); err != nil {
		log.Printf("[ERROR] update host: %v", err)
		rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "update host")
		return
	}

	rest.OkResponse(w)
}

func (s *Rest) deleteTarget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	if err := s.Config.DeleteHost(ctx, id); err != nil {
		log.Printf("[ERROR] delete host: %v", err)
		rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "delete host")
		return
	}

	rest.OkResponse(w)
}
