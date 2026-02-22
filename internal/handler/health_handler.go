package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping menangani request test connection ke aplikasi.
// GET /api/v1/ping
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "pong",
	})
}
