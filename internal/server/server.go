package server

import (
	"SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"SmartSpend/internal/server/handlers"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
	"strconv"
)

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	dbService := database.New() // This is of type database.Service
	database.RunMigrations(dbService.DB())

	userRepo := repository.NewUserRepository(dbService) // Pass the service

	serverHandler := &handlers.Server{
		Port:     port,
		Db:       dbService, // The database service
		UserRepo: userRepo,
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: serverHandler.RegisterRoutes(),
	}

	return srv
}
