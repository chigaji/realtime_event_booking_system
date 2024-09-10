package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, description, total_tickets, booked_tickets, event_date, created_at FROM events")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.Event

	for rows.Next() {
		var e models.Event
		if err = rows.Scan(&e.ID, &e.Name, &e.Description, &e.TotalTickets, &e.BookedTickets, &e.EventDate, &e.CreatedAt); err != nil {
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

	// See if this event is allready saved in the db
	var existingEvent models.Event
	err := s.db.QueryRow("SELECT id FROM events WHERE name = $1", event.Name).Scan(&existingEvent.ID)
	if err == nil {
		return fmt.Errorf("event already exists")
	}
	_, err = s.db.Exec("INSERT INTO events (name, description, total_tickets, event_date) VALUES($1, $2, $3, $4)", event.Name, event.Description, event.TotalTickets, event.EventDate)
	if err != nil {
		return err
	}

	//invalidate cache
	s.redis.Del(context.Background(), "events")
	return nil
}
