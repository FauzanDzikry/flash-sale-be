package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"flash-sale-be/internal/config"
	"flash-sale-be/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Jwt(cfg *config.Config, blacklist store.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format"})
			c.Abort()
			return
		}

		rawToken := parts[1]
		if blacklist != nil && blacklist.IsBlacklisted(rawToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token has been revoked"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTKey), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		if userID, ok := claims["user_id"].(string); ok {
			c.Set("user_id", userID)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}
		c.Set("token_raw", rawToken)
		if exp, ok := claims["exp"]; ok {
			switch v := exp.(type) {
			case float64:
				c.Set("token_exp", time.Unix(int64(v), 0))
			case int64:
				c.Set("token_exp", time.Unix(v, 0))
			}
		}
		c.Next()
	}
}
