package main

import (
	"log"
	"restaurant-system/internal/auth"
	"restaurant-system/internal/config"
	"restaurant-system/internal/database"
	"restaurant-system/internal/handlers"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"os"

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

	// Initialize GORM (for menu management)
	pgURL := os.Getenv("PG_URL")
	gdb, gerr := database.NewGorm(pgURL)
	if gerr != nil {
		log.Fatal("Failed to init GORM:", gerr)
	}

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
	// New grouped APIs
	adminAPI := handlers.NewAdminAPI()
	staffAPI := handlers.NewStaffAPI()
	customerAPI := handlers.NewCustomerAPI()

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

		// Menu (QR view remains)
		api.GET("/restaurant/:restaurant_id/table/:table_id/menu", menuAPI.GetQRMenu)
		// Menu management (GORM-backed)
		mm := handlers.NewMenuManagementAPI(gdb, hub)
		menuGroup := api.Group("/menu")
		{
			// categories
			menuGroup.POST("/categories", handlers.RequireAdminOrManager(), mm.CreateCategory)
			menuGroup.GET("/categories/:id", mm.GetCategory)
			menuGroup.GET("/categories", mm.ListCategories)
			menuGroup.PUT("/categories/:id", handlers.RequireAdminOrManager(), mm.UpdateCategory)
			menuGroup.DELETE("/categories/:id", handlers.RequireAdminOrManager(), mm.DeleteCategory)
			// items
			menuGroup.POST("/items", handlers.RequireAdminOrManager(), mm.CreateItem)
			menuGroup.GET("/items/:id", mm.GetItem)
			menuGroup.GET("/items", mm.ListItems)
			menuGroup.PUT("/items/:id", handlers.RequireAdminOrManager(), mm.UpdateItem)
			menuGroup.PATCH("/items/:id/availability", handlers.RequireStaff(), mm.UpdateAvailability)
			menuGroup.DELETE("/items/:id", handlers.RequireAdminOrManager(), mm.DeleteItem)
			// variants
			menuGroup.POST("/items/:id/variants", handlers.RequireAdminOrManager(), mm.CreateVariant)
			menuGroup.PUT("/variants/:id", handlers.RequireAdminOrManager(), mm.UpdateVariant)
			menuGroup.DELETE("/variants/:id", handlers.RequireAdminOrManager(), mm.DeleteVariant)
			// addons
			menuGroup.POST("/items/:id/addons", handlers.RequireAdminOrManager(), mm.CreateAddon)
			menuGroup.PUT("/addons/:id", handlers.RequireAdminOrManager(), mm.UpdateAddon)
			menuGroup.DELETE("/addons/:id", handlers.RequireAdminOrManager(), mm.DeleteAddon)
		}
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

		// Authenticated user routes (customer-facing)
		customer := api.Group("")
		customer.Use(auth.RequireAnyRole("customer", "waiter", "chef", "cashier", "admin"))
		{
			// profile
			customer.GET("/me", customerAPI.Me)
			customer.PATCH("/me", customerAPI.UpdateMe)
			// loyalty & promo
			customer.GET("/loyalty", customerAPI.GetLoyalty)
			customer.POST("/loyalty/redeem", customerAPI.RedeemPoints)
			customer.POST("/promo/apply", customerAPI.ApplyPromo)
			// waitlist (customer create/update own entry)
			customer.POST("/branches/:branchId/waitlist", customerAPI.CreateWaitlist)
			customer.PATCH("/branches/:branchId/waitlist/:entryId", customerAPI.UpdateWaitlist)
			// favorites, reviews, notifications
			customer.POST("/favorites", favoritesAPI.AddFavorite)
			customer.DELETE("/favorites/:menu_item_id", favoritesAPI.RemoveFavorite)
			customer.POST("/reviews", reviewsAPI.CreateReview)
			customer.POST("/notifications/subscribe", notificationsAPI.Subscribe)
			customer.DELETE("/notifications/:id", notificationsAPI.Unsubscribe)
			customer.GET("/notifications/account/:account_id", notificationsAPI.ListForAccount)
		}

		// Staff-facing
		staff := api.Group("/staff")
		staff.Use(auth.RequireAnyRole("waiter", "chef", "host", "cashier", "manager", "admin"))
		{
			// table states
			staff.PATCH("/branches/:branchId/tables/:tableId/state", staffAPI.UpdateTableState)
			// assignments
			staff.POST("/branches/:branchId/tables/:tableId/assign", staffAPI.AssignWaiterToTable)
			staff.POST("/branches/:branchId/orders/:orderId/assign-chef", staffAPI.AssignChefToOrder)
			// order lifecycle extensions
			staff.POST("/branches/:branchId/orders/:orderId/split", staffAPI.SplitOrder)
			staff.POST("/branches/:branchId/orders/merge", staffAPI.MergeOrders)
			staff.POST("/branches/:branchId/orders/:orderId/tip", staffAPI.AddTip)
		}

		// Kitchen (admin/staff scoped)
		kitchen := api.Group("/kitchen")
		kitchen.Use(auth.RequireAnyRole("chef", "manager", "admin"))
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

		// Admin-facing
		mgmt := api.Group("/admin")
		mgmt.Use(auth.RequireAnyRole("admin", "manager"))
		{
			// inventory
			mgmt.GET("/branches/:branchId/inventory", adminAPI.ListInventory)
			mgmt.POST("/branches/:branchId/inventory", adminAPI.CreateInventoryItem)
			mgmt.PATCH("/branches/:branchId/inventory/:itemId/adjust", adminAPI.AdjustInventory)
			mgmt.POST("/branches/:branchId/inventory/recipes", adminAPI.LinkRecipe)
			// reports
			mgmt.GET("/reports/sales", adminAPI.SalesReport)
			mgmt.GET("/reports/popular-items", adminAPI.PopularItemsReport)
			mgmt.GET("/reports/customers", adminAPI.CustomersReport)
			mgmt.GET("/reports/operations", adminAPI.OperationsReport)
			// branches
			mgmt.GET("/branches", adminAPI.ListBranches)
			mgmt.POST("/branches", adminAPI.CreateBranch)
			mgmt.PATCH("/branches/:branchId", adminAPI.UpdateBranch)
			// staff
			mgmt.GET("/staff", adminAPI.ListStaff)
			mgmt.POST("/staff", adminAPI.CreateStaff)
			mgmt.PATCH("/staff/:staffId", adminAPI.UpdateStaff)
		}
	}

	// Websocket endpoint (simple)
	router.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Websocket endpoint for menu updates (shares same hub; clients filter by type)
	router.GET("/ws/menu-updates", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Serve static files (optional; Next.js runs separately)
	router.Static("/static", "./web/static")

	log.Println("Starting restaurant system server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
