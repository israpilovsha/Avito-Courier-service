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
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/gateway/order"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/metrics"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/middleware"
	"github.com/IBM/sarama"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/db"
	deliveryHandler "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/handler"
	deliveryRepo "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/repository"
	deliveryUsecase "github.com/Avito-courses/course-go-avito-israpilovsha/internal/delivery/usecase"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/server/config"
	"github.com/Avito-courses/course-go-avito-israpilovsha/internal/worker"
	"github.com/Avito-courses/course-go-avito-israpilovsha/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	completeService := deliveryUsecase.NewCompleteService(
		deliveryRepository,
		courierRepository,
	)

	courierH := courierHandler.NewHandler(courierService, log)
	deliveryH := deliveryHandler.NewHandler(deliveryService, log)

	metrics.Register()

	r := mux.NewRouter()
	r.Use(middleware.MetricsAndLogging(log))

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	courierHandler.RegisterCourierRoutes(r, courierH)
	deliveryHandler.RegisterDeliveryRoutes(r, deliveryH)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go deliveryService.StartAutoRelease(ctx, cfg.Delivery.TickerInterval)

	// Kafka consumer
	if cfg.Kafka.Enabled {
		kafkaCfg := sarama.NewConfig()
		kafkaCfg.Version = sarama.V2_6_0_0
		kafkaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

		group, err := sarama.NewConsumerGroup(
			cfg.Kafka.Brokers,
			cfg.Kafka.GroupID,
			kafkaCfg,
		)
		if err != nil {
			log.Fatal("Kafka consumer group init failed", zap.Error(err))
		}

		orderGateway := order.NewHTTPGateway(cfg.OrderServiceHost)

		processor := worker.NewOrderEventProcessor(
			deliveryService,
			deliveryService,
			completeService,
			orderGateway,
		)

		handler := worker.NewOrderConsumer(processor, log)

		consumer := worker.NewKafkaConsumer(
			group,
			cfg.Kafka.Topic,
			handler,
			log,
		)

		consumer.Start(ctx)
		log.Info("Kafka consumer started")
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)
}
