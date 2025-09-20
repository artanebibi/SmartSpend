package handlers

import (
	db "SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"SmartSpend/internal/server/middleware"
	"SmartSpend/internal/service/application"
	"SmartSpend/internal/service/domain"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	database db.Service = db.New()

	userRepository        repository.IUserRepository        = repository.NewUserRepository(database)
	transactionRepository repository.ITransactionRepository = repository.NewTransactionRepository(database)
	categoryRepository    repository.ICategoryRepository    = repository.NewCategoryRepository(database)
	statisticsRepository  repository.IStatisticsRepository  = repository.NewStatisticsRepository(database)
	savingRepository      repository.ISavingRepository      = repository.NewSavingRepository(database)

	userService        domain.IUserService        = domain.NewUserService(userRepository)
	jwtService         domain.IJWTService         = domain.NewJWTService()
	tokenService       domain.ITokenService       = domain.NewTokenService()
	transactionService domain.ITransactionService = domain.NewTransactionService(transactionRepository)
	categoryService    domain.ICategoryService    = domain.NewCategoryService(categoryRepository)
	statisticsService  domain.IStatisticsService  = domain.NewStatisticsService(statisticsRepository)
	geminiService      domain.IGeminiService      = domain.NewGeminiService()

	applicationUserService        application.IUserAppService                = application.NewUserAppService(userService)
	applicationTransactionService application.IApplicationTransactionService = application.NewApplicationTransactionService(transactionRepository)
	applicationSavingService      application.IApplicationSavingService      = application.NewApplicationSavingService(savingRepository)
)

func parseFlexibleTime(timeStr string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	timeStr = strings.ReplaceAll(timeStr, " ", "+")
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

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
	authBasePath := "/api/auth"
	userBasePath := "/api/user"
	tokenBasePath := "/api/token"
	transactionBasePath := "/api/transaction"
	categoryBasePath := "/api/category"
	currencyBasePath := "/api/currency"
	statisticsBasePath := "/api/statistics"
	savingsBasePath := "/api/saving"

	r.GET("/health", s.healthHandler)

	r.GET("/websocket", s.websocketHandler)

	signIn := r.Group(authBasePath)
	{
		signIn.POST("/google", s.GoogleAuth)
		signIn.POST("/apple", s.AppleSignIn)
	}

	logOut := r.Group(authBasePath, middleware.AuthMiddleware())
	{
		logOut.POST("/logout", s.Logout)
	}

	token := r.Group(tokenBasePath)
	{
		token.POST("/", s.RotateAccessToken)
	}

	user := r.Group(userBasePath, middleware.AuthMiddleware())
	{
		user.GET("/me", s.GetUserData)
		user.GET("/balances", s.GetUserBalances)
		user.PATCH("/update", s.UpdateUserInformation)
	}

	transaction := r.Group(transactionBasePath, middleware.AuthMiddleware())
	{
		transaction.GET("", s.GetAllTransactions)
		transaction.GET("/:id", s.GetTransactionByID)
		transaction.POST("", s.SaveTransaction)
		transaction.PATCH("/:id", s.UpdateTransaction)
		transaction.DELETE("/:id", s.DeleteTransaction)
		transaction.POST("/receipt", s.SaveFromReceipt)
	}

	category := r.Group(categoryBasePath, middleware.AuthMiddleware())
	{
		category.GET("", s.GetAllCategories)
	}

	currency := r.Group(currencyBasePath, middleware.AuthMiddleware())
	{
		currency.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"data": []string{"MKD", "USD", "EUR"},
			})
		})
	}

	statistics := r.Group(statisticsBasePath, middleware.AuthMiddleware())
	{
		statistics.GET("/pie", s.Pie)                                   // sum of money spent (grouped by categories) - includes categories user has used
		statistics.GET("/monthly", s.Monthly)                           // sum of money spent per-month (all categories) - includes only months user has made transactions on
		statistics.GET("/total-spent", s.TotalSpentOnExpensesAndIncome) // sum of money used on expenses and income alone
		statistics.GET("/average", s.Average)                           // average price spent for expense and added for incomes
	}

	saving := r.Group(savingsBasePath, middleware.AuthMiddleware())
	{
		saving.GET("", s.GetAllSavings)
		saving.GET("/:id", s.GetSavingByID)
		saving.POST("", s.SaveSaving)
		saving.PATCH("/:id", s.UpdateSaving)
		saving.DELETE("/:id", s.DeleteTransaction)
	}

	return r
}
