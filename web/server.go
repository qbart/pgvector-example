package web

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"qbart/pgvector/ui"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	EmbeddableResources
	render Renderer
}

func (s *Server) Run(args []string) int {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/health/live"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		w.Write(s.EmbeddableResources.Favicon)
	})
	r.Get("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(s.EmbeddableResources.Style)
	})
	r.Get("/static/logo.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		w.Write(s.EmbeddableResources.Logo)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		s.render.HTML(w, r, ui.Dashboard())
	})

	server := &http.Server{Addr: ":3000", Handler: r}
	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-notifyCtx.Done()
	server.Shutdown(context.Background())

	return 0
}
