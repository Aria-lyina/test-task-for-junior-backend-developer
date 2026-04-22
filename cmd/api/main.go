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

	"github.com/robfig/cron/v3"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/repeat"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Репозитории
	taskRepo := postgresrepo.NewTaskRepository(pool)
	periodRepo := postgresrepo.NewPeriodRepository(pool)
	repeatTaskRepo := postgresrepo.NewRepeatTaskRepository(pool)

	// Usecases
	taskUsecase := task.NewService(taskRepo)
	repeatUsecase := repeat.NewService(periodRepo, repeatTaskRepo, taskRepo)

	// Хендлеры
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	repeatHandler := httphandlers.NewRepeatTaskHandler(repeatUsecase)
	docsHandler := swaggerdocs.NewHandler()

	router := transporthttp.NewRouter(taskHandler, repeatHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Планировщик с московским временем
	mskLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		logger.Warn("failed to load Moscow timezone, falling back to UTC+3", "error", err)
		mskLocation = time.FixedZone("MSK", 3*60*60)
	}

	c := cron.New(cron.WithLocation(mskLocation))

	// Каждый день в 00:00 по московскому времени
	_, err = c.AddFunc("0 0 * * *", func() {
	// _, err = c.AddFunc("12 5 * * *", func() {
		// Таймаут на случай если генерация зависнет
		jobCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		count, err := repeatUsecase.GenerateAllTasks(jobCtx)
		if err != nil {
			logger.Error("failed to generate tasks", "error", err)
		} else {
			logger.Info("tasks generated", "count", count)
		}
	})
	if err != nil {
		logger.Error("failed to add cron job", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown: сначала cron, потом HTTP сервер
	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received")

		// Останавливаем cron — ждём завершения текущих джобов
		cronCtx := c.Stop()
		select {
		case <-cronCtx.Done():
			logger.Info("cron stopped")
		case <-time.After(30 * time.Second):
			logger.Warn("cron stop timed out")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		} else {
			logger.Info("http server stopped")
		}
	}()

	// Запускаем cron только после успешной инициализации
	c.Start()
	logger.Info("cron scheduler started")

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func loadConfig() (config, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return config{}, fmt.Errorf("DATABASE_DSN is required")
	}

	return config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN: dsn,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}