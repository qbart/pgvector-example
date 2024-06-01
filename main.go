package main

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"qbart/pgvector/internal/documents"
	"qbart/pgvector/ui"
	"qbart/pgvector/web"
	"syscall"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	//go:embed assets/images/logo.svg
	logo []byte

	//go:embed ui/style.css
	style []byte

	//go:embed db/migrations/*.sql
	migrations embed.FS
)

func main() {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(os.Getenv("DATABASE_URL"))))
	db := bun.NewDB(sqldb, pgdialect.New())
	err := db.Ping()
	if err != nil {
		slog.Error("db ping", "message", err.Error())
		panic(err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("goose set dialect", "message", err.Error())
		panic(err)
	}
	if err := goose.Up(sqldb, "db/migrations"); err != nil {
		slog.Error("goose up", "message", err.Error())
		panic(err)
	}

	openAI, err := documents.NewOpenAI(os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		slog.Error("openai init failed", "message", err.Error())
		panic(err)
	}
	documentsHandler := &documents.HttpHandler{DB: db, OpenAI: openAI}

	r := chi.NewRouter()
	renderer := &web.Renderer{}

	// NOTE: this demo application is not production-ready,
	// in order to keept it simple, security features are omitted.
	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/health/live"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		w.Write(logo)
	})
	r.Get("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(style)
	})
	r.Get("/static/logo.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		w.Write(logo)
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		renderer.HTML(w, r, ui.Dashboard())
	})
	r.Get("/documents", documentsHandler.SearchDocument)
	r.Post("/documents", documentsHandler.UploadDocument)

	server := &http.Server{Addr: ":3000", Handler: r}
	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server listen and serve", "message", err.Error())
			panic(err)
		}
	}()

	<-notifyCtx.Done()
	server.Shutdown(context.Background())
}
