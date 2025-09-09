package middleware

import (
	"fmt"
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
	"strings"
)

// Middleware holds all middleware components
type Middleware struct {
	Logger *zap.Logger
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(logger *zap.Logger) *Middleware {
	return &Middleware{
		Logger: logger,
	}
}

// CORS configures CORS middleware
func (m *Middleware) CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// RateLimit configures rate limiting middleware
func (m *Middleware) RateLimit() gin.HandlerFunc {
	// 100 requests per minute
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}
	
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	
	return func(c *gin.Context) {
		context, err := instance.Get(c.Request.Context(), c.ClientIP())
		if err != nil {
			m.Logger.Error("Rate limiting error", zap.Error(err))
			c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			return
		}
		
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		
		if context.Reached {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many requests"})
			return
		}
		
		c.Next()
	}
}

// RequestLogger configures request logging middleware
func (m *Middleware) RequestLogger() gin.HandlerFunc {
	return ginzap.Ginzap(m.Logger, time.RFC3339, true)
}

// Recovery handles panics and returns 500 errors
func (m *Middleware) Recovery() gin.HandlerFunc {
	return ginzap.RecoveryWithZap(m.Logger, true)
}

// JWT configures JWT authentication middleware
func (m *Middleware) JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "authorization header required"})
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token claims"})
			return
		}
		userID, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing user ID in token"})
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// RequestID adds a unique request ID to each request
func (m *Middleware) RequestID() gin.HandlerFunc {
	return requestid.New()
}
