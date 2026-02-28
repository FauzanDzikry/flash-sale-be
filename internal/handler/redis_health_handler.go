package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RedisHealthHandler struct {
	client *redis.Client
}

func NewRedisHealthHandler(client *redis.Client) *RedisHealthHandler {
	return &RedisHealthHandler{client: client}
}

// PingRedis handles test connection request to Redis.
// GET /api/v1/ping/redis
func (h *RedisHealthHandler) Ping(c *gin.Context) {
	if h.client == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Redis client not configured",
		})
		return
	}

	if err := h.client.Ping(c.Request.Context()).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to connect to Redis",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "redis pong",
	})
}

