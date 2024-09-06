package api

import (
	"database/sql"

	"github.com/chigaji/realtime_event_booking_system/internal/api/handlers"
	"github.com/chigaji/realtime_event_booking_system/internal/config"
	"github.com/chigaji/realtime_event_booking_system/internal/middleware"
	"github.com/chigaji/realtime_event_booking_system/internal/queue"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func SetupRoutes(db *sql.DB, redis *redis.Client, q *queue.Queue, logger *zap.Logger, cfg *config.Config) *mux.Router {

	r := mux.NewRouter()

	auth := handlers.NewAuthHandler(db)
	event := handlers.NewEventHandler(db, redis)
	booking := handlers.NewBookingHandler(db, redis, q)
	mw := middleware.NewMiddleware(redis, logger, cfg)

	r.HandleFunc("/home", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/register", auth.Register).Methods("POST")
	r.HandleFunc("/login", auth.Login).Methods("POST")
	r.HandleFunc("/events", event.CreateEvents).Methods("POST")
	r.HandleFunc("/events", mw.RateLimit(event.GetEvents)).Methods("GET")
	r.HandleFunc("/book", mw.Authenticate(mw.RateLimit(booking.BookTicket))).Methods("POST")

	return r
}
