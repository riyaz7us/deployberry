package handlers

import (
	"errors"
	"net/http"
	"os"
	"shared/repository"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginData struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Login(c *gin.Context) {
	var login LoginData
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Get the stored password from auth bucket
	db := repository.GetDB()
	var user repository.User
	err := db.Where("username = ?", login.Username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// After successful password verification, generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: login.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get JWT secret from environment first, then fallback to database
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		db = repository.GetDB()
		var jwtSecret repository.Config
		err = db.Where("key = ?", "jwt_secret").First(&jwtSecret).Error
		if err == nil {
			secretStr = jwtSecret.Value
		}
	}

	if secretStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get JWT secret"})
		return
	}

	tokenString, err := token.SignedString([]byte(secretStr))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func LoggedIn(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}
