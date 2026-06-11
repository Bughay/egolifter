package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Bughay/egolifter/internal/auth"
	"github.com/Bughay/egolifter/internal/config"
	"github.com/Bughay/egolifter/internal/nutrition"
	"github.com/Bughay/egolifter/internal/recipe"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	// Auth module: handler -> service -> repository
	jwtManager := auth.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiryHours, cfg.JWT.RefreshExpiryDays*24)
	userRepo := auth.NewUserRepository(pool)
	authSvc := auth.NewAuthService(userRepo, jwtManager)
	authHandler := auth.NewAuthHandler(authSvc)
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)

	// Nutrition module: handler -> service -> repository
	foodRepo := nutrition.NewFoodRepository(pool)
	nutritionSvc := nutrition.NewNutritionService(foodRepo)
	nutritionHandler := nutrition.NewNutritionHandler(nutritionSvc)
	nutritionHandler.RegisterRoutes(mux)

	// Meal endpoints (nutrition module, JWT-protected)
	mealRepo := nutrition.NewMealRepository(pool)
	mealSvc := nutrition.NewMealService(mealRepo)
	mealHandler := nutrition.NewMealHandler(mealSvc)
	mealHandler.RegisterRoutes(mux, jwtManager.Middleware)

	// Recipe module: handler -> service -> repository (JWT-protected)
	recipeRepo := recipe.NewRecipeRepository(pool)
	recipeSvc := recipe.NewRecipeService(recipeRepo)
	recipeHandler := recipe.NewRecipeHandler(recipeSvc)
	recipeHandler.RegisterRoutes(mux, jwtManager.Middleware)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
	}

	log.Printf("server listening on :%s", cfg.Server.Port)
	log.Fatal(srv.ListenAndServe())
}
