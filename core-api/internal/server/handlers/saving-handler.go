package handlers

import (
	"SmartSpend/internal/domain/dto"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) GetAllSavings(c *gin.Context) {
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
	savings := applicationSavingService.FindAll(userId, from, to)

	c.JSON(200, gin.H{"data": savings})
}

func (s *Server) GetSavingByID(c *gin.Context) {
	id := c.Param("id")
	idInteger, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	_, userId := getUserFromDatabase(c)
	saving, err := applicationSavingService.FindById(idInteger, userId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": saving})
	return
}

func (s *Server) SaveSaving(c *gin.Context) {
	var savingDto dto.SavingDto
	if err := c.ShouldBindJSON(&savingDto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, userId := getUserFromDatabase(c)
	err, message := applicationSavingService.CreateOrUpdate(&savingDto, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": message})
	} else {
		c.JSON(200, gin.H{
			"message": message,
		})
	}
	return
}

func (s *Server) UpdateSaving(c *gin.Context) {
	var savingDto dto.SavingDto
	if err := c.ShouldBindJSON(&savingDto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, userId := getUserFromDatabase(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	savingDto.ID = id
	err, message := applicationSavingService.CreateOrUpdate(&savingDto, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": message})
	} else {
		c.JSON(200, gin.H{
			"message": message,
		})
	}
	return
}

func (s *Server) DeleteSaving(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, userId := getUserFromDatabase(c)

	saving, err_ := applicationSavingService.FindById(id, userId)
	if err_ != nil {
		c.JSON(400, gin.H{"error": err_.Error()})
	}
	err := applicationSavingService.Delete(saving, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		c.JSON(200, gin.H{
			"message": "Saving deleted successfully",
		})
	}
	return
}
