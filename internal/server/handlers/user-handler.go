package handlers

import (
	"SmartSpend/internal/domain/dto"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getUserFromDatabase(c *gin.Context) (*dto.UserDto, uuid.UUID) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "need to include a Authorization header"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, _ := tokenService.DecodeAccessToken(tokenString)
	userIDStr, _ := claims["UserID"].(string)

	userID, _ := uuid.Parse(userIDStr)

	user, err := applicationUserService.FindById(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	}
	return user, userID
}

func (s *Server) GetUserData(c *gin.Context) {
	user, _ := getUserFromDatabase(c)

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})

	return
}

func (s *Server) GetUserBalances(c *gin.Context) {
	user, _ := getUserFromDatabase(c)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"balance":             user.Balance,
			"monthly_saving_goal": user.MonthlySavingGoal,
		},
	})
}

func (s *Server) UpdateUserInformation(c *gin.Context) {
	_, userID := getUserFromDatabase(c)

	var u dto.UpdateUserDto
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	err := applicationUserService.Update(userID, u)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User information successfully updated.",
	})
	return
}
