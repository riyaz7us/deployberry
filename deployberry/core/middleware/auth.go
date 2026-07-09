package middleware

import (
	"net/http"
	"os"
	"shared/repository"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip middleware for login route
		if c.Request.URL.Path == "/api/login" || c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		if len(jwtSecret) == 0 {
			db := repository.GetDB()
			var jwtConfig repository.Config
			if err := db.Where("key = ?", "jwt_secret").First(&jwtConfig).Error; err == nil && jwtConfig.Value != "" {
				jwtSecret = []byte(jwtConfig.Value)
				os.Setenv("JWT_SECRET", jwtConfig.Value)
			}
		}

		if len(jwtSecret) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication secret is not configured"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
