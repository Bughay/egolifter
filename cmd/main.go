// cmd/api/main.go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bughay/go-backend-layout/internal/auth"
	"github.com/Bughay/go-backend-layout/internal/config"
	"github.com/Bughay/go-backend-layout/internal/database"
	"github.com/Bughay/go-backend-layout/internal/handler"
	"github.com/Bughay/go-backend-layout/internal/middleware"
	"github.com/Bughay/go-backend-layout/internal/repository"
	"github.com/Bughay/go-backend-layout/internal/service"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// 1. Load env and initialize structured JSON logger
	_ = godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// 2. Load typed configuration (fails fast on missing env vars)
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 3. Initialize the PostgreSQL connection pool
	pool, err := database.NewPool(context.Background(), cfg.Database.DSN)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Wire dependencies (Dependency Injection, manually)
	jwtManager := auth.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiryHours, cfg.JWT.RefreshExpiryDays)

	userRepo := repository.NewUserRepository(pool)

	authSvc := service.NewAuthService(userRepo, jwtManager)

	authHandler := handler.NewAuthHandler(authSvc)

	authDir := "/home/lordmark1/Desktop/go_boilerplate/frontend/frontend/auth"
	apphubDir := "/home/lordmark1/Desktop/go_boilerplate/frontend/frontend/pickservice"

	apphubProtected := jwtManager.CookieMiddleware(
		http.StripPrefix("/apphub/", middleware.NoCacheMiddleware(http.FileServer(http.Dir(apphubDir)))),
	)
	// 5. Register routes using the Go 1.22+ enhanced ServeMux
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	// static routes
	mux.Handle("/auth/", http.StripPrefix("/auth/", middleware.NoCacheMiddleware(http.FileServer(http.Dir(authDir)))))
	mux.Handle("/apphub/", apphubProtected)

	// 6. Apply global middleware (outermost = last to execute for the request, first for the response)
	loggedMux := middleware.Logger(logger)(mux)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1:5501", "http://localhost:5501"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(loggedMux)
	// 7. Configure the server with strict timeouts — essential for production
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 8. Start server in a goroutine and listen for OS shutdown signals
	go func() {
		logger.Info("server starting", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Block until we receive SIGINT or SIGTERM (e.g., from `docker stop` or Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 9. Graceful shutdown: give in-flight requests 30 seconds to complete
	logger.Info("server shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shut down", "error", err)
	}
	logger.Info("server stopped")
}
