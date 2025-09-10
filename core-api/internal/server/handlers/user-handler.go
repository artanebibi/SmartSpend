package handlers

import (
	"SmartSpend/internal/domain/dto"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func getUserFromDatabase(c *gin.Context) (*dto.UserDto, string) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "need to include a Authorization header"})
		return nil, ""
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, _ := tokenService.DecodeAccessToken(tokenString)
	userID, ok := claims["user-id"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "debug": claims["user-id"].(string)})
		return nil, ""
	}

	user, err := applicationUserService.FindById(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return nil, ""
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
