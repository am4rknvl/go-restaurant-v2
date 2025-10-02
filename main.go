package main

import (
	"fmt"
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

	// Initialize GORM (for menu management and enterprise features)
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		host := getenvDefault("PG_HOST", "localhost")
		port := getenvDefault("PG_PORT", "5432")
		user := getenvDefault("PG_USER", "postgres")
		password := getenvDefault("PG_PASSWORD", "postgres")
		dbname := getenvDefault("PG_DATABASE", "restaurant")
		sslmode := getenvDefault("PG_SSLMODE", "disable")
		pgURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	}
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
	enterpriseAPI := handlers.NewEnterpriseAPI(gdb, hub)
	orderWSHandler := handlers.NewOrderWSHandler(hub)

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

		// Enterprise endpoints
		api.GET("/accounts/:id", enterpriseAPI.GetAccount)
		api.PUT("/accounts/:id", enterpriseAPI.UpdateAccount)
		api.POST("/accounts/:id/roles", enterpriseAPI.AssignRole)
		api.DELETE("/accounts/:id/roles/:role", enterpriseAPI.RemoveRole)

		api.GET("/inventory", enterpriseAPI.ListInventory)
		api.POST("/inventory", enterpriseAPI.CreateInventoryItem)
		api.PUT("/inventory/:id", enterpriseAPI.UpdateInventoryItem)
		api.PATCH("/inventory/:id/adjust", enterpriseAPI.AdjustInventory)

		api.POST("/tables/:table_id/assign-waiter", enterpriseAPI.AssignWaiterToTable)
		api.POST("/orders/:order_id/assign-chef", enterpriseAPI.AssignChefToOrder)
		api.GET("/staff/assignments", enterpriseAPI.ListStaffAssignments)

		api.POST("/orders/:id/split", enterpriseAPI.SplitOrder)
		api.POST("/orders/:id/merge", enterpriseAPI.MergeOrders)
		api.POST("/payments/:id/tip", enterpriseAPI.AddTip)

		api.POST("/discounts", enterpriseAPI.CreateDiscount)
		api.POST("/discounts/apply", enterpriseAPI.ApplyDiscount)
		api.GET("/accounts/:id/loyalty", enterpriseAPI.GetLoyalty)
		api.POST("/accounts/:id/loyalty/earn", enterpriseAPI.EarnLoyaltyPoints)

		api.GET("/reports/sales", enterpriseAPI.SalesReport)
		api.GET("/reports/popular-items", enterpriseAPI.PopularItemsReport)
		api.GET("/reports/customers/top", enterpriseAPI.TopCustomersReport)

		api.GET("/restaurants", enterpriseAPI.ListRestaurants)
		api.POST("/restaurants", enterpriseAPI.CreateRestaurant)
		api.PUT("/restaurants/:id", enterpriseAPI.UpdateRestaurant)

		api.PATCH("/tables/:id/state", enterpriseAPI.UpdateTableState)
		api.POST("/waitlist", enterpriseAPI.JoinWaitlist)
		api.GET("/waitlist", enterpriseAPI.ListWaitlist)

		// Enterprise APIs
		// User profiles & role management
		api.GET("/accounts/:id", auth.RequireAnyRole("admin", "manager"), enterpriseAPI.GetAccount)
		api.PUT("/accounts/:id", auth.RequireAnyRole("admin", "manager"), enterpriseAPI.UpdateAccount)
		api.POST("/accounts/:id/roles", auth.RequireAnyRole("admin"), enterpriseAPI.AssignRole)
		api.DELETE("/accounts/:id/roles/:role", auth.RequireAnyRole("admin"), enterpriseAPI.RemoveRole)

		// Inventory & stock management
		api.GET("/inventory", auth.RequireAnyRole("chef", "manager", "admin"), enterpriseAPI.ListInventory)
		api.POST("/inventory", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.CreateInventoryItem)
		api.PUT("/inventory/:id", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.UpdateInventoryItem)
		api.PATCH("/inventory/:id/adjust", auth.RequireAnyRole("chef", "manager", "admin"), enterpriseAPI.AdjustInventory)

		// Staff assignment
		api.POST("/tables/:table_id/assign-waiter", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.AssignWaiterToTable)
		api.POST("/orders/:order_id/assign-chef", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.AssignChefToOrder)
		api.GET("/staff/assignments", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.ListStaffAssignments)

		// Order lifecycle extensions
		api.POST("/orders/:id/split", auth.RequireAnyRole("waiter", "cashier", "manager", "admin"), enterpriseAPI.SplitOrder)
		api.POST("/orders/:id/merge", auth.RequireAnyRole("waiter", "cashier", "manager", "admin"), enterpriseAPI.MergeOrders)
		api.POST("/payments/:id/tip", auth.RequireAnyRole("customer", "cashier"), enterpriseAPI.AddTipToPayment)

		// Loyalty & discounts
		api.POST("/discounts", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.CreateDiscount)
		api.POST("/discounts/apply", enterpriseAPI.ApplyDiscount)
		api.GET("/accounts/:id/loyalty", enterpriseAPI.GetLoyaltyAccount)
		api.POST("/accounts/:id/loyalty/earn", auth.RequireAnyRole("cashier", "manager", "admin"), enterpriseAPI.EarnLoyaltyPoints)

		// Analytics & reporting
		api.GET("/reports/sales", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.SalesReport)
		api.GET("/reports/popular-items", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.PopularItemsReport)
		api.GET("/reports/customers/top", auth.RequireAnyRole("manager", "admin"), enterpriseAPI.TopCustomersReport)

		// Multi-restaurant / branch support
		api.GET("/restaurants", enterpriseAPI.ListRestaurants)
		api.POST("/restaurants", auth.RequireAnyRole("admin"), enterpriseAPI.CreateRestaurant)
		api.PUT("/restaurants/:id", auth.RequireAnyRole("admin"), enterpriseAPI.UpdateRestaurant)

		// Table state management & waitlist
		api.PATCH("/tables/:id/state", auth.RequireAnyRole("waiter", "host", "manager", "admin"), enterpriseAPI.UpdateTableState)
		api.POST("/waitlist", enterpriseAPI.JoinWaitlist)
		api.GET("/waitlist", enterpriseAPI.ListWaitlist)
	}

	// Websocket endpoint (simple)
	router.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Websocket endpoint for menu updates (shares same hub; clients filter by type)
	router.GET("/ws/menu-updates", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Websocket endpoint for order updates
	router.GET("/ws/orders", orderWSHandler.HandleOrderUpdates)

	// WebSocket endpoint for waitlist updates
	router.GET("/ws/waitlist", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c.Writer, c.Request)
	})

	// Serve static files (optional; Next.js runs separately)
	router.Static("/static", "./web/static")

	log.Println("Starting restaurant system server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
