package tasks

import (
	"github.com/google/uuid"
	"time"
)

// TaskStatus defines the possible states of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusScheduled TaskStatus = "SCHEDULED"
	StatusRunning   TaskStatus = "RUNNING"
	StatusCompleted TaskStatus = "COMPLETED"
	StatusFailed    TaskStatus = "FAILED"
)

type Task struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	Payload   []byte     `json:"playload"`
	Status    TaskStatus `json:"status"`
	ExecuteAt time.Time  `json:"execute_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewTask(taskType string, payload []byte, executeAt time.Time) *Task {
	return &Task{
		ID:        uuid.New(),
		Type:      taskType,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		ExecuteAt: executeAt,
	}
}
