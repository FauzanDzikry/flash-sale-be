package testutil

import (
	"context"
	"testing"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	redismod "github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupTestRedis(t *testing.T) (*redisclient.Client, func()) {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := redismod.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("failed to get redis endpoint: %v", err)
	}

	rdb := redisclient.NewClient(&redisclient.Options{
		Addr: endpoint,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}

	cleanup := func() {
		_ = rdb.Close()
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}

	return rdb, cleanup
}

func SetupTestRedisWithConnectionString(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	redisContainer, err := redismod.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}

	cleanup = func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}

	return connStr, cleanup
}

