package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/handler"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/service"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/server/config"
	"github.com/Avito-courses/course-go-avito-israpilovsha/pkg/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	log := logger.New()
	defer log.Sync()

	cfg := config.MustLoad()
	log.Info("Configuration loaded", zap.String("port", cfg.Port))

	database := db.New(cfg.Postgres.DSN(), log)
	defer database.Close()

	repo := repository.NewPostgresCourierRepository(database)
	svc := service.NewCourierService(repo)
	h := handler.NewHandler(svc)

	r := mux.NewRouter()
	handler.RegisterCourierRoutes(r, h)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("Server is starting...", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server shutdown failed", zap.Error(err))
	}

	log.Info("Server stopped gracefully")
}
