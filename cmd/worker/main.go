package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/student9876/go-dist-scheduler/internal/store"
	"github.com/student9876/go-dist-scheduler/internal/tasks"
)

func main() {
	// 1. Connect to Prostgres
	// we need this to fetch the actual data (payload) of the tasks
	dbPool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))

	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}
	defer dbPool.Close()
	taskStore := store.NewPostgresTaskStore(dbPool)

	// 2. Connect to RabbitMQ
	// wait for RabbitMQ to be ready
	var conn *amqp091.Connection
	rabbitURL := os.Getenv("RABBITMQ_URL")

	for i := 0; i < 30; i++ {
		conn, err = amqp091.Dial(rabbitURL)
		if err == nil {
			log.Println("Connected to RabbitMQ!")
			break
		}
		log.Printf("Waiting for RabbitMQ... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	// 3. decalrequeue
	// ensure queue exists
	q, _ := ch.QueueDeclare("tasks_queue", true, false, false, false, nil)

	// Start Consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Worker started. Waiting for tasks...")

	// 5. Processing loop
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			taskID := string(d.Body)
			log.Printf("Received Task ID: %s", taskID)

			ctx := context.Background()

			// A. Update status
			taskStore.UpdateStatus(ctx, taskID, tasks.StatusRunning)

			// B. Fetch task details
			task, err := taskStore.Get(ctx, taskID)
			if err != nil {
				log.Printf("Error fetching task: %v", err)
				d.Reject(false)
				continue
			}

			// C. execute the logic
			if err := executeTask(task); err != nil {
				log.Printf("task failed: %v", err)
				taskStore.UpdateStatus(ctx, taskID, tasks.StatusFailed)
				// d.Reject(true) // Optional: Requeue the task
				d.Reject(false)
			} else {
				log.Printf("Task finished successfully!")
				taskStore.UpdateStatus(ctx, taskID, tasks.StatusCompleted)
				d.Ack(false)
			}

		}
	}()
	<-forever
}

// executeTask simulates doing the actual work
func executeTask(t *tasks.Task) error {
	// This is where you would put your email sending logic, etc.
	fmt.Printf(">>> 🚀 EXECUTING TASK: Type=%s | Payload=%s\n", t.Type, string(t.Payload))
	time.Sleep(2 * time.Second) // Simulate work taking time
	return nil
}
