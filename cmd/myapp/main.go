package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	courierHandler "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/handler"
	courierRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/repository"
	courierUsecase "github.com/Avito-courses/course-go-avito-israpilovsha/internal/courier/usecase"

	deliveryHandler "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/handler"
	deliveryRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
	deliveryUsecase "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/usecase"
	orderGateway "github.com/Avito-courses/course-go-avito-israpilovsha/internal/gateway/order"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/worker"

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

	courierRepository := courierRepo.NewCourierRepository(database)
	deliveryRepository := deliveryRepo.NewDeliveryRepository(database)

	courierService := courierUsecase.NewCourierService(courierRepository)
	deliveryService := deliveryUsecase.NewDeliveryService(
		courierRepository,
		deliveryRepository,
	)

	courierH := courierHandler.NewHandler(courierService, log)
	deliveryH := deliveryHandler.NewHandler(deliveryService, log)

	r := mux.NewRouter()
	courierHandler.RegisterCourierRoutes(r, courierH)
	deliveryHandler.RegisterDeliveryRoutes(r, deliveryH)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go deliveryService.StartAutoRelease(ctx, cfg.Delivery.TickerInterval)
	log.Info("Background auto-release task started")

	orderGW := orderGateway.NewHTTPGateway(cfg.OrderServiceHost)
	orderFetcher := worker.NewOrderFetcher(orderGW, deliveryService, log)

	go orderFetcher.Run(ctx)
	log.Info("Background order fetcher started")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

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
