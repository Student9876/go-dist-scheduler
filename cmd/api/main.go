package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/student9876/go-dist-scheduler/internal/store"
	"github.com/student9876/go-dist-scheduler/internal/tasks"
)

// define request
type CreateTaskRequest struct {
	Type      string                 `json:"type" binding:"required"`
	Playload  map[string]interface{} `json:"payload" binding:"required"`
	ExecuteAt time.Time              `json:"execute_at" binding:"required"`
}

func main() {
	// setup dependencies
	dbPool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	taskStore := store.NewPostgresTaskStore(dbPool)
	schedulerStore := store.NewRedisScheduleStore(redisClient)

	// Initialize Gin
	r := gin.Default()

	// Define routes

	r.POST("/schedule", func(c *gin.Context) {
		// A. Parse and validate request
		var req CreateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// B. convert payload to bytes
		payloadBytes, err := json.Marshal(req.Playload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
			return
		}

		// C. create task object
		newTask := tasks.NewTask(req.Type, payloadBytes, req.ExecuteAt)

		// D. save to DB (using a timeout context)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := taskStore.Create(ctx, newTask); err != nil {
			log.Printf("DB Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save task"})
			return
		}

		// E. Save to Redis
		if err := schedulerStore.AddToSchedule(ctx, newTask.ID.String(), newTask.ExecuteAt); err != nil {
			log.Printf("Redis Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule task"})
			return
		}

		// F. success
		c.JSON(http.StatusOK, gin.H{
			"task_id": newTask.ID.String(),
			"status":  "scheduled",
		})
	})

	// Run server
	port := "8080"
	log.Printf("API Listening on port %s", port)
	r.Run(":" + port)
}
