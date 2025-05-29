package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scoreboard-api/internal"
	"scoreboard-api/internal/config"
	"scoreboard-api/internal/database"
	"scoreboard-api/internal/scoreboard"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file")
	}

	cfg := config.Load()
	logger, err := zap.NewDevelopment()
	validator := internal.NewValidator()
	if err != nil {
		panic(err)
	}
	err = database.MigrationUp(cfg.MigrationSource, cfg.DatabaseURL, logger)
	if err != nil {
		panic(err)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	service := scoreboard.NewService(logger, db)
	handler := scoreboard.NewHandler(validator, logger, service)

	mux := http.NewServeMux()

	// Handles GET /api/scoreboards and POST /api/scoreboards
	mux.HandleFunc("/api/scoreboards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListHandler(w, r)
		case http.MethodPost:
			handler.CreateHandler(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	// Handles GET /api/scoreboards/{id}, PUT /api/scoreboards/{id}, DELETE /api/scoreboards/{id}
	mux.HandleFunc("/api/scoreboards/", func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Path will be like "/api/scoreboards/some-id"
		// The individual handlers (GetHandler, UpdateHandler, DeleteHandler)
		// are responsible for extracting the ID from the path and validating it.
		switch r.Method {
		case http.MethodGet:
			handler.GetHandler(w, r)
		case http.MethodPut:
			handler.UpdateHandler(w, r)
		case http.MethodDelete:
			handler.DeleteHandler(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server", zap.String("address", "127.0.0.1:8080"))
		if err := server.ListenAndServe(); err != nil {
			logger.Error("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("Failed to shutdown server", zap.Error(err))
	}
	logger.Info("Server shutdown successfully")
}
