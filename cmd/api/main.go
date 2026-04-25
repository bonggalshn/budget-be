package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	authapi "github.com/bonggalshn/budget-be/api/v1/auth"
	"github.com/bonggalshn/budget-be/internal/auth"
	"github.com/bonggalshn/budget-be/internal/config"
	"github.com/bonggalshn/budget-be/internal/db"
	"github.com/bonggalshn/budget-be/internal/loginattempt"
	"github.com/bonggalshn/budget-be/internal/session"
	"github.com/bonggalshn/budget-be/internal/user"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	sessionRepo := session.NewRepository(pool)
	attemptRepo := loginattempt.NewRepository(pool)

	authService := auth.NewService(*cfg, userRepo, sessionRepo, attemptRepo)
	authHandler := auth.NewHandler(authService, *cfg)
	authMiddleware := auth.NewMiddleware(authService, *cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheck)

	authHandlerAPI := authapi.Routes(authHandler, authMiddleware)
	mux.Handle("/api/v1/auth/", authHandlerAPI)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	srv.Shutdown(context.Background())
	log.Println("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}