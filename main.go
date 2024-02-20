package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("please set environment variables")
	}

	if os.Getenv("BASIC_AUTH_USER") == "" || os.Getenv("BASIC_AUTH_PASS") == "" {
		log.Println("please set BASIC_AUTH_USER and BASIC_AUTH_PASS")
		os.Exit(1)
	}
}

func main() {
	r := chi.NewRouter()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 認証無し
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	// 認証が必要なグループ
	// 今回は例としてBasic認証を用いる
	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin area", map[string]string{os.Getenv("BASIC_AUTH_USER"): os.Getenv("BASIC_AUTH_PASS")}))
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin page"))
		})
	})

	srv := &http.Server{
		Addr:    ":3333",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println("server error", err)
		}
	}()
	log.Println("Server is ready to handle requests at :3333")

	// graceful shutdown
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Failed to gracefully shutdown the server", err)
	}
	log.Println("Server shutdown")
}
