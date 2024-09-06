package main

import (
	"log"
	"net/http"

	"github.com/chigaji/realtime_event_booking_system/internal/api"
	"github.com/chigaji/realtime_event_booking_system/internal/config"
	"github.com/chigaji/realtime_event_booking_system/internal/database"
	"github.com/chigaji/realtime_event_booking_system/internal/queue"
	"github.com/chigaji/realtime_event_booking_system/internal/redis"
	"github.com/chigaji/realtime_event_booking_system/pkg/logger"
	"go.uber.org/zap"
)

// initialie logger

func main() {

	//initialize logger
	logger.Init()
	defer logger.Sync()

	cfg, err := config.LoadConfig()

	if err != nil {
		logger.Log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.Init(cfg.Redis)
	if err != nil {
		logger.Log.Info("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize queue
	q, err := queue.Init(cfg.Queue)
	if err != nil {
		logger.Log.Fatal("Failed to initialize queue", zap.Error(err))
	}
	defer q.Close()

	// Start the booking worker
	go q.StartBookingWorker(db, redisClient)

	// Initialize and start the HTTP server
	router := api.SetupRoutes(db, redisClient, q, logger.Log, cfg)
	logger.Log.Info("Server starting", zap.String("address", cfg.Server.Address))
	log.Fatal(http.ListenAndServe(cfg.Server.Address, router))

}
