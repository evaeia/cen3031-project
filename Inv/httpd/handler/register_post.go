package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	//current encrypt
	"golang.org/x/crypto/bcrypt"
	/*/ salting
	        "crypto/rand"
	        "crypto/sha256"
	        "encoding/base64"
	  // encrypting
	        "crypto/aes"
	        "crypto/cipher"
	        "crypto/rand"
	        "encoding/base64"
	        "fmt"
	        "io"*/)

func RegisterPost(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the request body.
		var request RegisterRequest
		if err := c.BindJSON(&request); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if the user exists.
		var count int64
		if result := DB.Table("accounts").Where("username = ?", request.Username).Count(&count); result.Error == nil && count > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}

		// Check if the password is correct.
		if request.Password != request.PasswordConfirmation {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Password and password confirmation do not match"})
			return
		}

		// Create a new user object with the provided username and password.
		newUser := Account{
			Username: request.Username,
			Password: HashAndSalt([]byte(request.Password)), //replaced with hash and salt password,
		}

		//check if container is empty
		var maxLocID int64
		if maxLocID = GetMaxLocID(DB); maxLocID == -1 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max location ID"})
			return
		}

		newContainer := Container{
			LocID:    int(maxLocID) + 1,
			Name:     newUser.Username + "'s container",
			ParentID: 0, // Assuming it's a top-level container.
			User:     newUser.Username,
		}

		var token = GenerateToken()
		// Create a new session
		session := DeviceSession{
			Username: newUser.Username,
			Token:    token,
			LastUsed: time.Now(),
			// Optionally add device identification here if needed
		}

		// Save the new session
		if result := DB.Create(&session); result.Error != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		// Start a new transaction to ensure atomicity.
		tx := DB.Begin()

		// Create the new user and container objects in the database.
		if result := tx.Table("accounts").Create(&newUser); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		if result := tx.Table("containers").Create(&newContainer); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
			return
		}

		// Update the new user's RootLoc to the LocID of the new container.
		if result := tx.Table("accounts").Where("username = ?", newUser.Username).Update("rootLoc", newContainer.LocID); result.Error != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user's RootLoc"})
			return
		}

		// Commit the transaction.
		tx.Commit()

		// Return the token to the user.
		response := LoginResponse{Token: token, RootLoc: newContainer.LocID}
		c.JSON(http.StatusOK, response)
	}
}

// Hash and Salt password
func HashAndSalt(password []byte) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	// Hash the password using the salt
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		println(err)
		return ""
	}
	// Convert the hash to a string and return it
	return string(hash)
}

func GetMaxLocID(DB *gorm.DB) int64 {
	var maxLocID int64

	// Check if the "containers" table is empty
	var count int64
	DB.Table("containers").Count(&count)
	if count == 0 {
		// Return 0 if the table is empty
		return 0
	}

	err := DB.Table("containers").Select("MAX(LocID)").Row().Scan(&maxLocID)
	if err != nil {
		return -1
	}

	return maxLocID

}
