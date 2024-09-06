package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/chigaji/realtime_event_booking_system/pkg/utils"
	"github.com/go-redis/redis/v8"
)

type EventHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewEventHandler(db *sql.DB, redis *redis.Client) *EventHandler {

	return &EventHandler{
		db:    db,
		redis: redis,
	}
}

func (h *EventHandler) CreateEvents(w http.ResponseWriter, r *http.Request) {
	// ctx := context.Background()

	// get eventdata from request
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// cachedEvents, err := h.redis.Get(ctx, "events").Result()

	// try to get events from the cache first
	// if err != nil {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write([]byte(cachedEvents))
	// 	return
	// }

	// See if this event is allready saved in the db
	var existingEvent models.Event
	err := h.db.QueryRow("SELECT id FROM events WHERE name = $1", event.Name).Scan(&existingEvent.ID)
	if err == nil {
		http.Error(w, "Event Already Exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// save event to the database
	_, err = h.db.Exec("INSERT INTO events (name,description, total_tickets, event_date) VALUES ($1, $2, $3, $4)", event.Name, event.Description, event.TotalTickets, event.EventDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": utils.DBCreateEventResp})
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	cachedEvents, err := h.redis.Get(ctx, "events").Result()

	// try to get events from the cache first
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedEvents))
		return
	}

	// if no in cache get from the database
	rows, err := h.db.QueryContext(ctx, "SELECT id, name, description, total_tickets, booked_tickets, booked_tickets, event_date, created_at FROM events")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var events []models.Event

	for rows.Next() {
		var e models.Event

		if err = rows.Scan(&e.ID, &e.Name, &e.Description, &e.TotalTickets, &e.BookedTickets, &e.EventDate, &e.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, e)
	}

	// cache the result
	eventJSON, _ := json.Marshal(events)
	h.redis.Set(ctx, "events", eventJSON, time.Minute*5)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
