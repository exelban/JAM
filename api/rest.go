package api

import (
	"crypto/subtle"
	"github.com/exelban/cheks/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pkgz/rest"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:generate moq -out mock_test.go . monitor

type monitor interface {
	Status() map[string]types.StatusType
	Services() []types.Service
}

type Auth struct {
	Enabled  bool
	Username string
	Password string
}

type Rest struct {
	Monitor monitor
	Version string
	FS      fs.FS
	Auth    Auth
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
	router.Use(s.basicAuth)

	router.HandleFunc("/admin", s.admin)
	router.HandleFunc("/admin/*", s.admin)

	router.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		rest.JsonResponse(w, s.Monitor.Services())
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

func (s *Rest) admin(w http.ResponseWriter, r *http.Request) {
	p := strings.Replace(r.URL.Path, "/admin/", "/admin/dist/", 1)
	rp := strings.Replace(r.URL.RawPath, "/admin/", "/admin/dist/", 1)

	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Path = p
	r2.URL.RawPath = rp

	http.FileServer(http.FS(s.FS)).ServeHTTP(w, r2)
}
