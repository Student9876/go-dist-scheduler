package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go" // RabbitMQ client
	"github.com/redis/go-redis/v9"
	"github.com/student9876/go-dist-scheduler/internal/store"
)

func main() {
	// 1. connect to redis
	// we need to check fo due tasks

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	// 2. connect to rabbitmq
	var conn *amqp091.Connection
	var err error

	rabbitmqURL := os.Getenv("RABBITMQ_URL")

	for i := 0; i < 20; i++ {
		conn, err = amqp091.Dial(rabbitmqURL)
		if err == nil {
			log.Printf("Connected to RabbitMQ")
			break
		}
		log.Printf("Waiting for RabbitMQ... (%v)", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// 3. Open a channel
	// In RabbitMQ, you communicate via "Channels", not the connection directly.
	ch, _ := conn.Channel()
	defer ch.Close()

	// 4. Declare the Queue
	// Ensure the queue "tasks_queue" exists. If it doesn't, RabbitMQ created it.
	q, err := ch.QueueDeclare(
		"tasks_queue", //name
		true,          //durable (messages save to disk so they survive server restarts)
		false,         //delete when unused
		false,         //exclusive
		false,         // no-wait
		nil,           //arguments
	)

	if err != nil {
		log.Fatal(err)
	}

	// 5. Initialize the redis store
	schedulerStore := store.NewRedisScheduleStore(redisClient)
	ctx := context.Background()

	log.Println("Scheduler started. Polling Redis every 2 seconds...")

	// 6. The polling loop (heartbeat)
	ticker := time.NewTicker(2 * time.Second)

	for range ticker.C {
		// A. Ask redis if anything is due right now
		dueTasks, err := schedulerStore.PullDueTasks(ctx)
		if err != nil {
			log.Printf("Error pulling due tasks from redis store: %v", err)
		}

		// B. If we found tasks, push them to RabbitMQ
		for _, taskID := range dueTasks {
			err = ch.Publish(
				"",     //exchange (default)
				q.Name, //routing key (queue name)
				false,  //mandatory
				false,  //immediate
				amqp091.Publishing{
					ContentType:  "text/plain",
					Body:         []byte(taskID),     // we send only the task ID
					DeliveryMode: amqp091.Persistent, // save to disk
				},
			)
			if err != nil {
				log.Printf("Failed to publish task %s: %v", taskID, err)
			} else {
				log.Printf("Moved task %s from Redis -> RabbitMQ", taskID)
			}
		}
	}
}
