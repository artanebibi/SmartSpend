package handlers

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) Pie(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	_, userId := getUserFromDatabase(c)

	from, err := parseFlexibleTime(fromStr)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	to, err := parseFlexibleTime(toStr)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
	}

	percentages, totalExpenses, totalIncome, err := statisticsService.FindPercentageSpentPerCategory(userId, from, to)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"data": []gin.H{
			{
				"statistics":     percentages,
				"total_expenses": totalExpenses,
				"total_income":   totalIncome,
				"from":           from,
				"to":             to,
			},
		},
	})
	return

}
