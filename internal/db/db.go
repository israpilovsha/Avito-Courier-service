package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Database struct {
	Pool *pgxpool.Pool
	log  *zap.SugaredLogger
}

func New(databaseURL string, logger *zap.SugaredLogger) *Database {
	var pool *pgxpool.Pool
	var err error

	const maxRetries = 5
	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, databaseURL)
		if err == nil {
			// Пинг базы
			if errPing := pool.Ping(ctx); errPing == nil {
				logger.Infof("Database connected successfully")
				return &Database{Pool: pool, log: logger}
			} else {
				err = errPing
			}
		}

		logger.Warnf("Database ping failed (attempt %d/%d): %v", i, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	logger.Fatalw("Failed to connect to database after multiple attempts", "error", err)
	return nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.log.Info("Database connection closed")
	}
}
