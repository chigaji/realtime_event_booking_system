package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/go-redis/redis/v8"
)

type EventService struct {
	db    *sql.DB
	redis *redis.Client
}

func NewEventService(db *sql.DB, redis *redis.Client) *EventService {
	return &EventService{db: db, redis: redis}
}

func (s *EventService) GetEvents(ctx context.Context) ([]models.Event, error) {

	// get events from the cached first
	cachedEvents, err := s.redis.Get(ctx, "events").Result()
	if err == nil {
		var events []models.Event
		err = json.Unmarshal([]byte(cachedEvents), &events)
		if err != nil {
			return nil, err
		}
	}

	//if not get events from the db
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, total_tickets, booked_tickets FROM events")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.Event

	for rows.Next() {
		var e models.Event
		if err = rows.Scan(&e.ID, &e.TotalTickets, &e.BookedTickets); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	// cache the results
	eventJSON, _ := json.Marshal(events)
	s.redis.Set(ctx, "events", eventJSON, time.Minute*5)

	return events, nil
}

func (s *EventService) CreateEvent(event *models.Event) error {

	_, err := s.db.Exec("INSERT INTO events (name, total_tickets, booked_tickets) VALUES($1, $2, 0)", event.Name, event.TotalTickets)
	if err != nil {
		return err
	}

	//invalidate cache
	s.redis.Del(context.Background(), "events")
	return nil
}
