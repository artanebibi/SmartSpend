package handlers

import (
	db "SmartSpend/internal/database"
	"SmartSpend/internal/repository"
	"SmartSpend/internal/server/middleware"
	"SmartSpend/internal/service/application"
	"SmartSpend/internal/service/domain"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	database                      db.Service                                 = db.New()
	userRepository                repository.IUserRepository                 = repository.NewUserRepository(database)
	userService                   domain.IUserService                        = domain.NewUserService(userRepository)
	applicationUserService        application.IUserAppService                = application.NewUserAppService(userService)
	jwtService                    domain.IJWTService                         = domain.NewJWTService()
	tokenService                  domain.ITokenService                       = domain.NewTokenService()
	transactionRepository         repository.ITransactionRepository          = repository.NewTransactionRepository(database)
	transactionService            domain.ITransactionService                 = domain.NewTransactionService(transactionRepository)
	applicationTransactionService application.IApplicationTransactionService = application.NewApplicationTransactionService(transactionRepository)
	categoryRepository            repository.ICategoryRepository             = repository.NewCategoryRepository(database)
	categoryService               domain.ICategoryService                    = domain.NewCategoryService(categoryRepository)
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
	authBasePath := "/api/auth"
	userBasePath := "/api/user"
	tokenBasePath := "/api/token"
	transactionBasePath := "/api/transaction"
	categoryBasePath := "/api/category"
	currencyBasePath := "/api/currency"

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
	return r
}
