package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/chigaji/realtime_event_booking_system/internal/queue"
	"github.com/go-redis/redis/v8"
)

type BookingService struct {
	db    *sql.DB
	redis *redis.Client
	queue *queue.Queue
}

func NewBookingService(db *sql.DB, redis *redis.Client, queue *queue.Queue) *BookingService {
	return &BookingService{db: db, redis: redis, queue: queue}
}

func (s *BookingService) RequestBooking(booking *models.Booking) error {
	return s.queue.PublishBookingRequest(*booking)
}

// func (s *BookingService) ProcessBooking() {
// 	// return queue.ProcessBooking(s.db, s.redis, *booking)
// 	s.queue.StartBookingWorker(s.db, s.redis)
// }

func (s *BookingService) ProcessBooking(ctx context.Context, booking *models.Booking) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var availableTickets int
	err = tx.QueryRowContext(ctx, "SELECT total_tickets - booked_tickets FROM events WHERE id = $1 FOR UPDATE", booking.EventID).Scan(&availableTickets)
	if err != nil {
		return err
	}

	if availableTickets <= 0 {
		return errors.New("no tickets available")
	}

	_, err = tx.ExecContext(ctx, "UPDATE events SET booked_tickets = booked_tickets + 1 WHERE id = $1", booking.EventID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO bookings (user_id, event_id) VALUES ($1, $2)", booking.UserID, booking.EventID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Invalidate cache
	s.redis.Del(ctx, "events")

	return nil
}
