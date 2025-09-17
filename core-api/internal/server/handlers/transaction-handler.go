package handlers

import (
	"SmartSpend/internal/domain/dto"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
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

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (s *Server) SaveFromReceipt(c *gin.Context) {
	_, userId := getUserFromDatabase(c)

	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get image file"})
		return
	}
	defer file.Close()

	var imgBuffer bytes.Buffer
	teeReader := io.TeeReader(file, &imgBuffer)

	img, _, err := image.Decode(teeReader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image"})
		return
	}

	base64Image, err := imageToBase64(img)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image to Base64"})
		return
	}

	tx, err := geminiService.SendToGemini("", base64Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("Gemini service error: %w", err).Error()})
		return
	}

	tx.OwnerId = userId
	tx.DateMade = time.Now()

	log.Println("Gemini Transaction:", tx)

	c.JSON(http.StatusOK, tx)
}
