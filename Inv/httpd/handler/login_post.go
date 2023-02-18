package handler

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type User struct {
	Username string `gorm:"primaryKey"`
	Password string
	Token    string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginPost(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the request body.
		var request LoginRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if the user exists.
		var user User
		if result := DB.Where("username = ?", request.Username).First(&user); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Check if the password is correct.
		if user.Password != request.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Generate a token and save it to the database.
		token := generateToken()
		user.Token = token
		if result := DB.Save(&user); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}

		// Return the token to the user.
		response := LoginResponse{Token: token}
		c.JSON(http.StatusOK, response)
	}
}

func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
