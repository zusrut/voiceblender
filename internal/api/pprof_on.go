//go:build pprof

package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

func registerPprof(r *chi.Mux, log interface{ Info(string, ...any) }) {
	log.Info("pprof enabled")
	r.Get("/debug/pprof/", http.HandlerFunc(pprof.Index))
	r.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Post("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Get("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.Get("/debug/pprof/{profile}", http.HandlerFunc(pprof.Index))
}
