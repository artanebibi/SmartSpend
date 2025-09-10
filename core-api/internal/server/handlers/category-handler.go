package handlers

import "github.com/gin-gonic/gin"

func (s *Server) GetAllCategories(c *gin.Context) {
	categories := categoryService.FindAll()

	if len(categories) == 0 {
		c.JSON(200, gin.H{
			"data": "No categories found",
		})
		return
	} else {
		c.JSON(200, gin.H{
			"data": categories,
		})
	}
	return
}
