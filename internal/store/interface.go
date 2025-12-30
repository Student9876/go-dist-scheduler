package store

import (
	"context"
	"time"

	"github.com/student9876/go-dist-scheduler/internal/tasks"
)

// how we interact with the peranent storage
type TaskStore interface {
	// saves a new task to the DB
	Create(ctx context.Context, task *tasks.Task) error

	// retrieves a task by its ID
	Get(ctx context.Context, id string) (*tasks.Task, error)

	// changes the status of a task
	UpdateStatus(ctx context.Context, id string, status tasks.TaskStatus) error
}

// how we interact with the fast scheduler storage (redis)
type SchedulerStore interface {
	// adds a task id with timestamp
	AddToSchedule(ctx context.Context, taskID string, executeAt time.Time) error

	// PullDueTaks retreives all task IDs that are due for execution
	// It should also remove them from the schedule so they aren't picked twice
	PullDueTasks(ctx context.Context) ([]string, error)
}
