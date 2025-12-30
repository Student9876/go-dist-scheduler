package store

import (
	"context"
	"fmt"
	// "time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/student9876/go-dist-scheduler/internal/tasks"
)

type PostgresTaskStore struct {
	db *pgxpool.Pool
}

// creates a new instance
func NewPostgresTaskStore(db *pgxpool.Pool) *PostgresTaskStore {
	return &PostgresTaskStore{db: db}
}

// inserts a new task
func (s *PostgresTaskStore) Create(ctx context.Context, t *tasks.Task) error {
	query := `
		INSERT INTO tasks (id, type, payload, status, execute_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	// We use the pool to execute the query
	_, err := s.db.Exec(ctx, query,
		t.ID,
		t.Type,
		t.Payload,
		t.Status,
		t.ExecuteAt,
		t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// fetches a task by ID
func (s *PostgresTaskStore) Get(ctx context.Context, id string) (*tasks.Task, error) {
	query := `
		SELECT * 
		FROM tasks
		WHERE id = $1
	`
	var t tasks.Task

	err := s.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.Type,
		&t.Payload,
		&t.Status,
		&t.ExecuteAt,
		&t.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &t, nil
}

// modifies the status of a task
func (s *PostgresTaskStore) UpdateStatus(ctx context.Context, id string, status tasks.TaskStatus) error {
	query := `
		UPDATE tasks
		SET status = $2
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, id, status)

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}
