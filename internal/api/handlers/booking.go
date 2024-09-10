package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/chigaji/realtime_event_booking_system/internal/queue"
	"github.com/chigaji/realtime_event_booking_system/pkg/validator"
	"github.com/go-redis/redis/v8"
)

type BookingHandler struct {
	// redis *redis.Client
	db    *sql.DB
	redis *redis.Client
	queue *queue.Queue
	// bookingService *services.BookingService
}

func NewBookingHandler(db *sql.DB, redis *redis.Client, queue *queue.Queue) *BookingHandler {
	return &BookingHandler{
		db:    db,
		redis: redis,
		queue: queue,
	}
}

func (h *BookingHandler) BookTicket(w http.ResponseWriter, r *http.Request) {

	var booking models.Booking

	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validator.Validate(booking); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// checkRateLimit(booking.UserID)
	// implement rate limiting
	if !h.checkRateLimit(booking.UserID) {
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	//publish booking to queue
	if err := h.queue.PublishBookingRequest(booking); err != nil {
		http.Error(w, "Failed to process Booking", http.StatusInternalServerError)
		return
	}
	// ToDO: call here to start processing the booking
	// h.queue.StartBookingWorker(h.db, h.redis)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Booking request accepted and queued for processing"})
}

func (h *BookingHandler) checkRateLimit(userID int) bool {
	ctx := context.Background()

	key := fmt.Sprintf("rate_limit:%d", userID)

	n, err := h.redis.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Error Increamenting rate limit: %v", err)
		return false
	}
	if n == 1 {
		h.redis.Expire(ctx, key, time.Minute)
	}
	return n <= 5 //allow 5 request per minute
}
