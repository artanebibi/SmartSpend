package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) getAndParseTime(c *gin.Context) (time.Time, time.Time, error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("missing 'from' or 'to' query parameters")
	}

	from, err := parseFlexibleTime(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	to, err := parseFlexibleTime(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return from, to, nil
}

func (s *Server) Pie(c *gin.Context) {
	_, userId := getUserFromDatabase(c)
	from, to, err := s.getAndParseTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	percentages, totalExpenses, totalIncome, err := statisticsService.FindPercentageSpentPerCategory(userId, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
}

func (s *Server) Monthly(c *gin.Context) {
	_, userId := getUserFromDatabase(c)
	from, to, err := s.getAndParseTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	spentPerMonth, err := statisticsService.FindTotalSpentPerMonth(userId, from, to)
	totalExpense, totalIncome, err := statisticsService.FindTotalIncomeAndExpense(userId, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"statistics":     spentPerMonth,
				"total_expenses": totalExpense,
				"total_income":   totalIncome,
				"from":           from,
				"to":             to,
			},
		},
	})
	return
}

func (s *Server) TotalSpentOnExpensesAndIncome(c *gin.Context) {
	_, userId := getUserFromDatabase(c)
	from, to, err := s.getAndParseTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	totalExpense, totalIncome, err := statisticsService.FindTotalIncomeAndExpense(userId, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"total_expenses": totalExpense,
				"total_income":   totalIncome,
				"from":           from,
				"to":             to,
			},
		},
	})
	return
}

func (s *Server) Average(c *gin.Context) {
	_, userId := getUserFromDatabase(c)
	from, to, err := s.getAndParseTime(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	averageExpense, averageIncome, err := statisticsService.FindTotalIncomeAndExpense(userId, from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{
			{
				"average_expense": averageExpense,
				"average_income":  averageIncome,
				"from":            from,
				"to":              to,
			},
		},
	})
	return
}
