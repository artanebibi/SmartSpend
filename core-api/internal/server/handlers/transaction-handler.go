package handlers

import (
	"SmartSpend/internal/domain/dto"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetAllTransactions(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	var err error

	if fromStr == "" {
		from = time.Time{}
	} else {
		from, err = parseFlexibleTime(fromStr)
		if err != nil {
			log.Printf("Failed to parse 'from': %s, error: %v", fromStr, err)
			c.JSON(400, gin.H{"error": "invalid 'from' date format"})
			return
		}
	}

	if toStr == "" {
		to = time.Now()
	} else {
		to, err = parseFlexibleTime(toStr)
		if err != nil {
			log.Printf("Failed to parse 'to': %s, error: %v", toStr, err)
			c.JSON(400, gin.H{"error": "invalid 'to' date format"})
			return
		}
	}

	if !from.IsZero() && !to.IsZero() && from.After(to) {
		c.JSON(400, gin.H{"error": "'from' date cannot be after 'to' date"})
		return
	}

	_, userId := getUserFromDatabase(c)
	transactions := applicationTransactionService.FindAll(userId, from, to)

	c.JSON(200, gin.H{"data": transactions})
}

func (s *Server) GetTransactionByID(c *gin.Context) {
	id := c.Param("id")
	idInteger, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	_, userId := getUserFromDatabase(c)
	transaction, err := applicationTransactionService.FindById(idInteger, userId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": transaction})
	return
}

func (s *Server) SaveTransaction(c *gin.Context) {
	var t dto.TransactionDto
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, userId := getUserFromDatabase(c)
	err, message := applicationTransactionService.CreateOrUpdate(&t, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": message})
	} else {
		c.JSON(200, gin.H{
			"message": message,
		})
	}
	return
}

func (s *Server) UpdateTransaction(c *gin.Context) {
	var t dto.TransactionDto
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, userId := getUserFromDatabase(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	t.ID = id
	err, message := applicationTransactionService.CreateOrUpdate(&t, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": message})
	} else {
		c.JSON(200, gin.H{
			"message": message,
		})
	}
	return
}

func (s *Server) DeleteTransaction(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, userId := getUserFromDatabase(c)

	t, err_ := applicationTransactionService.FindById(id, userId)
	if err_ != nil {
		c.JSON(400, gin.H{"error": err_.Error()})
	}
	err := applicationTransactionService.Delete(t, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		c.JSON(200, gin.H{
			"message": "Transaction deleted successfully",
		})
	}
	return
}

func (s *Server) SaveFromReceipt(c *gin.Context) {
	resp, err := http.Post("http://localhost:5000/ocr", c.Request.Header.Get("Content-Type"), c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	log.Println(string(body))
	c.JSON(200, gin.H{
		"response from ocr": string(body),
	})
	return
}
