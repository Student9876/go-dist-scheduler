package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisScheduleStore struct {
	client *redis.Client
}

func NewRedisScheduleStore(client *redis.Client) *RedisScheduleStore {
	return &RedisScheduleStore{client: client}
}

// adds the task ID to a ZSET sorted by execution time
func (r *RedisScheduleStore) AddToSchedule(ctx context.Context, taskID string, executeAt time.Time) error {
	// adds a member (taskID) with a score (unix timestamp)
	err := r.client.ZAdd(ctx, "task_schedule", redis.Z{
		Score:  float64(executeAt.Unix()),
		Member: taskID,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to schedule task in redis: %w", err)
	}
	return nil
}

// PullDueTasks fetches tasks that are due and removes them
func (r *RedisScheduleStore) PullDueTasks(ctx context.Context) ([]string, error) {
	now := float64(time.Now().Unix())

	// 1. Fetch tasks
	tasks, err := r.client.ZRangeByScore(ctx, "task_schedule", &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch due tasks: %w", err)
	}

	if len(tasks) == 0 {
		return []string{}, nil
	}

	// --- FIX START: Convert []string to []interface{} ---
	// Redis needs generic interfaces to delete multiple items at once
	members := make([]interface{}, len(tasks))
	for i, v := range tasks {
		members[i] = v
	}
	// --- FIX END ---

	// 2. Remove them (Pass the converted slice)
	err = r.client.ZRem(ctx, "task_schedule", members...).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to remove tasks from schedule: %w", err)
	}

	return tasks, nil
}
