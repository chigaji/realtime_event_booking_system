package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/chigaji/realtime_event_booking_system/internal/config"
	"github.com/chigaji/realtime_event_booking_system/internal/models"
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func Init(cfg config.Queueconfig) (*Queue, error) {

	conn, err := amqp.Dial(cfg.Address)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	// q, err := ch.QueueDeclare(
	// 	"booking_requests",
	// 	true,
	// 	false,
	// 	false,
	// 	false,
	// 	nil,
	// )
	_, err = ch.QueueDeclare(
		"booking_requests",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Queue{conn: conn, ch: ch}, nil

}
func (q *Queue) Close() {
	q.ch.Close()
	q.conn.Close()
}

func (q *Queue) PublishBookingRequest(booking models.Booking) error {
	body, err := json.Marshal(booking)
	if err != nil {
		return err
	}

	return q.ch.Publish(
		"",
		"booking_requests",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (q *Queue) StartBookingWorker(db *sql.DB, redisClient *redis.Client) {

	msgs, err := q.ch.Consume(
		"booking_requests",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		var booking models.Booking
		err := json.Unmarshal(msg.Body, &booking)

		if err != nil {
			log.Printf("Error Unmarshaling request: %v", err)
			continue
		}
		err = ProcessBooking(db, redisClient, booking)
		if err != nil {
			log.Printf("Error Processing the booking: %v", err)
		}
	}
}

func ProcessBooking(db *sql.DB, redisClient *redis.Client, booking models.Booking) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check if tickets are available and book
	var availableTickets int
	err = tx.QueryRow("SELECT total_tickets - booked_tickets FROM events WHERE id = $1 FOR UPDATE", booking.EventID).Scan(&availableTickets)
	if err != nil {
		fmt.Println("error 1")
		return err
	}

	if availableTickets <= 0 {
		return errors.New("no available tickets")
	}

	// book the ticket
	_, err = tx.Exec("UPDATE events SET booked_tickets = booked_tickets +$1 WHERE id = $2", booking.Quantity, booking.EventID)
	if err != nil {
		fmt.Println("error 2")
		return err
	}

	// insert booking record
	err = tx.QueryRow("INSERT INTO bookings (user_id, event_id, quantity) VALUES ($1, $2, $3) RETURNING id", booking.UserID, booking.EventID, booking.Quantity).Scan(&booking.ID)
	if err != nil {
		fmt.Println("error 3")
		return err
	}

	//commit the transaction

	if err = tx.Commit(); err != nil {
		return err
	}

	//update catch

	ctx := context.Background()
	redisClient.Del(ctx, "events")
	return nil
}
