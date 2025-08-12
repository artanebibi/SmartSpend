package server

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/server/handlers"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
	"strconv"
	"time"
)

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &handlers.Server{
		Port: port,
		Db:   database.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.Port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
