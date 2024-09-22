package api

import (
	"net/http"
)

// Middleware type for middleware functions
type Middleware func(http.Handler) http.Handler

// Router struct to hold our routes and middleware
type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
	patterns    []string
}

// NewRouter creates and returns a new App with an initialized ServeMux and middleware slice
func NewRouter(middlewares ...Middleware) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		middlewares: middlewares,
	}
}

// Use adds middlewares to the chain
func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// HandleFunc registers a handler function for a specific route, applying all middleware.
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, handler)
}

// Handle registers a handler for a specific route, applying all middleware.
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.patterns = append(r.patterns, pattern)
	finalHandler := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		finalHandler = r.middlewares[i](finalHandler)
	}
	r.mux.Handle(pattern, finalHandler)
}
