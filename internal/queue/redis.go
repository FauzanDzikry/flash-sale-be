package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	CheckoutQueueKey     = "checkout_queue"
	DefaultBlockDuration = 5 * time.Second
)

var ErrEmptyQueue = errors.New("queue is empty")

type CheckoutJob struct {
	JobID      string    `json:"job_id"`
	UserID     string    `json:"user_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

type Queue interface {
	EnqueueCheckout(ctx context.Context, job CheckoutJob) error
	DequeueCheckout(ctx context.Context) (*CheckoutJob, error)
}

type redisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(client *redis.Client) Queue {
	return &redisQueue{client: client, key: CheckoutQueueKey}
}

func (q *redisQueue) EnqueueCheckout(ctx context.Context, job CheckoutJob) error {
	if job.JobID == "" {
		job.JobID = uuid.New().String()
	}
	job.EnqueuedAt = time.Now()
	b, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, q.key, b).Err()
}

func (q *redisQueue) DequeueCheckout(ctx context.Context) (*CheckoutJob, error) {
	val, err := q.client.BRPop(ctx, DefaultBlockDuration, q.key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrEmptyQueue
		}
		return nil, err
	}
	if len(val) < 2 {
		return nil, ErrEmptyQueue
	}
	var job CheckoutJob
	if err := json.Unmarshal([]byte(val[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
