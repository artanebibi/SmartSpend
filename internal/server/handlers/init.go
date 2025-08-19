package handlers

import (
	db "SmartSpend/internal/database" // Add an alias to avoid conflict
	"SmartSpend/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	Port     int
	Db       db.Service // Use the alias here
	UserRepo repository.IUserRepository
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/health", s.healthHandler)

	r.GET("/websocket", s.websocketHandler)

	signInBasePath := "/api/auth/signin"

	r.POST(signInBasePath+"/google", s.GoogleSignIn)
	r.POST(signInBasePath+"/apple", s.AppleSignIn)

	return r
}
