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

func (r *RedisScheduleStore) PullDueTasks(ctx context.Context) ([]string, error) {
	now := float64(time.Now().Unix())

	// fetch tasks with score from -infinity to NOW

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

	// remove them from the set so that other pollers don't grab them
	// This contains race condition
	err = r.client.ZRem(ctx, "tasks_schedule", tasks).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to remove tasks from schedule: %w", err)
	}
	return tasks, nil
}
