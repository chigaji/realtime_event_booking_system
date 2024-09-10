package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/chigaji/realtime_event_booking_system/internal/services"
	"github.com/chigaji/realtime_event_booking_system/pkg/utils"
	"github.com/go-redis/redis/v8"
)

type EventHandler struct {
	db           *sql.DB
	redis        *redis.Client
	eventService *services.EventService
}

func NewEventHandler(db *sql.DB, redis *redis.Client, eventService *services.EventService) *EventHandler {

	return &EventHandler{
		db:           db,
		redis:        redis,
		eventService: eventService,
	}
}

func (h *EventHandler) CreateEvents(w http.ResponseWriter, r *http.Request) {
	// ctx := context.Background()

	// get eventdata from request
	// fmt.Println("before")
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.eventService.CreateEvent(&event)
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
	if err == nil {
		fmt.Println("getting cached events")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedEvents))
		return
	}

	events, err := h.eventService.GetEvents(h.redis.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// // cache the result
	eventJSON, _ := json.Marshal(events)
	h.redis.Set(h.redis.Context(), "events", eventJSON, time.Minute*5)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
