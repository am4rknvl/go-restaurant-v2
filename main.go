package main

import (
	"log"
	"restaurant-system/internal/auth"
	"restaurant-system/internal/config"
	"restaurant-system/internal/database"
	"restaurant-system/internal/handlers"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}
	config.Load()

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize services (use available SQL services)
	orderService := services.NewOrderSQLService(db.Conn())
	paymentService := services.NewPaymentSQLService(db.Conn())
	accountService := services.NewAccountService(db)
	authService := services.NewAuthService(db)
	authBasicService := services.NewAuthBasicService(db.Conn())
	menuService := services.NewMenuSQLService(db.Conn())
	tableService := services.NewTableService(db.Conn())
	sessionService := services.NewSessionService(db.Conn())
	reservationService := services.NewReservationService(db.Conn())
	notificationService := services.NewNotificationService(db.Conn())

	// WebSocket hub for real-time updates
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	orderAPI := handlers.NewOrderAPI(orderService, hub)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	accountHandler := handlers.NewAccountHandler(accountService)
	authHandler := handlers.NewAuthHandler(authService)
	authBasicHandler := handlers.NewAuthBasicHandler(authBasicService)
	menuAPI := handlers.NewMenuAPI(menuService)
	menuAdmin := handlers.NewMenuAdminAPI(menuService)
	categoriesAPI := handlers.NewCategoriesAPI(menuService)
	favoritesAPI := handlers.NewFavoritesAPI(menuService)
	reviewsAPI := handlers.NewReviewsAPI(menuService)
	tablesAPI := handlers.NewTablesAPI(tableService)
	sessionsAPI := handlers.NewSessionsAPI(sessionService)
	kitchenAPI := handlers.NewKitchenAPI(orderService)
	reservationsAPI := handlers.NewReservationsAPI(reservationService)
	notificationsAPI := handlers.NewNotificationsAPI(notificationService)

	// Setup router
	router := gin.Default()

	// Share services in context
	router.Use(func(c *gin.Context) {
		var ps services.PaymentService = paymentService
		c.Set("paymentService", ps)
		c.Next()
	})

	// Enable CORS for frontend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (OTP + Basic)
		authGroup := api.Group("/auth")
		{
			// Optional OTP
			authGroup.POST("/request-otp", authHandler.RequestOTP)
			authGroup.POST("/verify-otp", authHandler.VerifyOTP)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
			// Basic Auth
			authGroup.POST("/signup", authBasicHandler.Signup)
			authGroup.POST("/signin", authBasicHandler.Signin)
		}

		// Order routes
		orders := api.Group("/orders")
		{
			orders.POST("", orderAPI.CreateOrder)
			orders.POST("/sync", orderAPI.SyncOrders)
			orders.GET("/:id", orderAPI.GetOrder)
			orders.GET("", orderAPI.ListOrders)
			orders.GET("/customer/:customer_id", orderAPI.ListOrdersByCustomer)
			orders.POST("/:id/reorder", orderAPI.Reorder)
			orders.PUT("/:id/status", orderAPI.UpdateOrderStatus)
			orders.PUT("/:id/eta", orderAPI.SetETA)
		}

		// Menu / categories
		api.GET("/restaurant/:restaurant_id/table/:table_id/menu", menuAPI.GetQRMenu)
		// Protect admin routes
		admin := api.Group("")
		admin.Use(auth.RequireAuth(""))
		menuAdminGroup := admin.Group("/menu")
		{
			menuAdminGroup.POST("/item", menuAdmin.CreateItem)
			menuAdminGroup.PUT("/item/:id", menuAdmin.UpdateItem)
			menuAdminGroup.DELETE("/item/:id", menuAdmin.DeleteItem)
		}
		cats := admin.Group("/categories")
		{
			cats.POST("", categoriesAPI.CreateCategory)
			cats.GET("", categoriesAPI.ListCategories)
			cats.PUT("/:id", categoriesAPI.UpdateCategory)
			cats.DELETE("/:id", categoriesAPI.DeleteCategory)
		}

		// Tables & sessions
		tables := admin.Group("/tables")
		{
			tables.POST("", tablesAPI.CreateTable)
			tables.GET("", tablesAPI.ListTables)
			tables.GET("/:id", tablesAPI.GetTable)
		}
		sessions := admin.Group("/sessions")
		{
			sessions.POST("", sessionsAPI.StartSession)
			sessions.GET("/:id", sessionsAPI.GetSession)
			sessions.PUT("/:id/close", sessionsAPI.CloseSession)
		}

		// Reservations
		reservations := api.Group("/reservations")
		{
			reservations.POST("", reservationsAPI.CreateReservation)
			reservations.GET("", reservationsAPI.ListReservations)
			reservations.PUT("/:id", reservationsAPI.UpdateReservation)
			reservations.DELETE("/:id", reservationsAPI.CancelReservation)
		}

		// Authenticated user routes
		user := api.Group("")
		user.Use(auth.RequireAuth(""))
		{
			user.POST("/favorites", favoritesAPI.AddFavorite)
			user.DELETE("/favorites/:menu_item_id", favoritesAPI.RemoveFavorite)
			user.POST("/reviews", reviewsAPI.CreateReview)
			user.POST("/notifications/subscribe", notificationsAPI.Subscribe)
			user.DELETE("/notifications/:id", notificationsAPI.Unsubscribe)
			user.GET("/notifications/account/:account_id", notificationsAPI.ListForAccount)
		}

		// Kitchen
		kitchen := admin.Group("/kitchen")
		{
			kitchen.GET("/orders", kitchenAPI.ListPending)
			kitchen.PUT("/orders/:id/status", kitchenAPI.UpdateStatus)
			kitchen.POST("/notifications/send", notificationsAPI.SendNotification)
		}

		// Payment routes
		payments := api.Group("/payments")
		{
			payments.POST("", paymentHandler.CreatePayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.POST("/:id/refund", paymentHandler.RequestRefund)
			payments.POST("/partial", paymentHandler.ApplyPartialPayment)
			payments.POST("/notify/telebirr", handlers.TelebirrNotifyHandler)
			payments.POST("/notify/chapa", handlers.TelebirrNotifyHandler)
			payments.POST("/notify/mpesa", handlers.TelebirrNotifyHandler)
		}

		// Account routes
		accounts := api.Group("/accounts")
		{
			accounts.GET("/:id/balance", accountHandler.GetAccountBalance)
			accounts.POST("", accountHandler.CreateAccount)
		}

		// Kitchen routes removed for now

		// WebSocket route removed for now
	}

	// Websocket endpoint (simple)
	router.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Serve static files (optional; Next.js runs separately)
	router.Static("/static", "./web/static")

	log.Println("Starting restaurant system server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
