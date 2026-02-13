package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/repository"
	"shortURL/internal/repository/memory"
	"shortURL/internal/repository/postgres"
	"shortURL/internal/service"
)

func main() {
	// загрузка конфигураций
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Storage type: %s", cfg.StorageType)
	log.Printf("Base URL: %s", cfg.BaseURL)

	// инициализируем в зависимости от типа хранения
	var repo repository.URLRepository
	var cleanup func()

	switch cfg.StorageType {
	case "memory":
		log.Println("Using in-memory storage")
		repo = memory.NewMemoryRepository()
		cleanup = func() {
			repo.Close()
		}

	case "postgres":
		log.Println("Connecting to PostgreSQL")
		pgRepo, err := postgres.NewPostgresRepository(cfg.PostgresConnectionString())
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}

		// инициализируем схему
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := pgRepo.InitSchema(ctx); err != nil {
			cancel()
			log.Fatalf("Failed to initialize schema: %v", err)
		}
		cancel()

		log.Println("PostgreSQL connected and schema initialized")
		repo = pgRepo
		cleanup = func() {
			log.Println("Closing PostgreSQL connection")
			repo.Close()
		}

	default:
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	defer cleanup()

	// инициализация юрлсервиса
	urlService := service.NewURLService(repo)

	// инициализация хендлера
	urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)

	mux := handler.SetupRoutes(urlHandler)

	// Создание сервера
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// старт сервера в горутине
	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	// даем серверу 30 сек чтобы выключится

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
