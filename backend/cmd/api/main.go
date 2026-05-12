package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"queryforge/backend/internal/config"
	"queryforge/backend/internal/db"
	"queryforge/backend/internal/handlers"
	appmw "queryforge/backend/internal/middleware"
	"queryforge/backend/internal/repository"
	"queryforge/backend/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		healthcheck(cfg.BackendPort)
		return
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.PostgresDSN())
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.RunMigrations(ctx, pool, "migrations", logger); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.StorageDir, 0o750); err != nil {
		logger.Error("create storage dir", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	workspaceRepo := repository.NewWorkspaceRepository(pool)
	historyRepo := repository.NewHistoryRepository(pool)
	schemaSvc := services.NewSchemaService()
	authSvc := services.NewAuthService(cfg, userRepo, tokenRepo)
	workspaceSvc := services.NewWorkspaceService(cfg, workspaceRepo, schemaSvc)
	aiClient := services.NewAIClient(cfg.AIServiceURL, cfg.AIRequestTimeout)
	querySvc := services.NewQueryService(workspaceRepo, historyRepo, schemaSvc, aiClient)

	authHandler := handlers.NewAuthHandler(authSvc)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceSvc, schemaSvc)
	queryHandler := handlers.NewQueryHandler(querySvc)
	aiHandler := handlers.NewAIHandler(aiClient)

	r := chi.NewRouter()
	r.Use(appmw.Recoverer(logger))
	r.Use(appmw.RequestLogger(logger))
	r.Use(appmw.CORS(cfg.FrontendOrigin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmw.Auth(authSvc))
			r.Get("/workspaces", workspaceHandler.List)
			r.Post("/workspaces", workspaceHandler.Create)
			r.Get("/workspaces/{id}", workspaceHandler.Get)
			r.Delete("/workspaces/{id}", workspaceHandler.Delete)
			r.Post("/workspaces/{id}/upload", workspaceHandler.Upload)
			r.Get("/workspaces/{id}/schema", workspaceHandler.Schema)
			r.Post("/workspaces/{id}/query/generate", queryHandler.Generate)
			r.Post("/workspaces/{id}/query/execute", queryHandler.Execute)
			r.Get("/workspaces/{id}/history", queryHandler.ListHistory)
			r.Get("/history/{historyId}", queryHandler.GetHistory)
			r.Get("/ai/health", aiHandler.Health)
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.BackendPort,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      cfg.AIRequestTimeout + 10*time.Second,
	}
	logger.Info("backend listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func healthcheck(port string) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
