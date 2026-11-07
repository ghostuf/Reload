package main

import (
	"log"
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/config"
	"github.com/ghostuf/reload/internal/handlers"
	"github.com/ghostuf/reload/internal/repository"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
)

// corsMiddleware allows credentialed requests from a single known origin.
// Browsers reject cookies on a response with "Access-Control-Allow-Origin: *",
// so the origin is echoed back only when it matches the configured one.
func corsMiddleware(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// The allowed origin varies per request, so caches must key on it.
		c.Header("Vary", "Origin")

		if c.GetHeader("Origin") == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	tokenManager := auth.NewTokenManager(cfg.JWTAccessSecret)
	cookieOpts := auth.CookieOptions{
		Domain: cfg.CookieDomain,
		Secure: cfg.CookieSecure,
	}

	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, tokenManager)
	authHandler := handlers.NewAuthHandler(authService, cookieOpts)

	reviewRepo := repository.NewReviewRepository(db)

	gameRepo := repository.NewGameRepository(db)
	gameService := services.NewGameService(gameRepo, reviewRepo)
	gameHandler := handlers.NewGameHandler(gameService)

	cartRepo := repository.NewCartRepository(db)
	cartService := services.NewCartService(cartRepo, gameRepo)
	cartHandler := handlers.NewCartHandler(cartService)

	wishlistRepo := repository.NewWishlistRepository(db)
	wishlistService := services.NewWishlistService(wishlistRepo, gameRepo)
	wishlistHandler := handlers.NewWishlistHandler(wishlistService)

	checkoutService := services.NewCheckoutService(db)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderService)

	libraryRepo := repository.NewLibraryRepository(db)
	libraryService := services.NewLibraryService(libraryRepo)
	libraryHandler := handlers.NewLibraryHandler(libraryService)

	reviewService := services.NewReviewService(reviewRepo, libraryRepo, gameRepo, userRepo)
	reviewHandler := handlers.NewReviewHandler(reviewService)

	router := gin.Default()
	router.Use(corsMiddleware(cfg.FrontendOrigin))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/games", gameHandler.ListGames)
	router.GET("/games/:id", gameHandler.GetGame)

	// Grouped under /auth to match the path the refresh cookie is scoped to,
	// so it is never sent along with ordinary API requests.
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.Refresh)
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.GET("/me", auth.RequireAuth(tokenManager), authHandler.Me)
	}

	cartRoutes := router.Group("/cart", auth.RequireAuth(tokenManager))
	{
		cartRoutes.GET("", cartHandler.GetCart)
		cartRoutes.POST("", cartHandler.AddToCart)
		cartRoutes.DELETE("/:game_id", cartHandler.RemoveFromCart)
	}

	wishlistRoutes := router.Group("/wishlist", auth.RequireAuth(tokenManager))
	{
		wishlistRoutes.GET("", wishlistHandler.GetWishlist)
		wishlistRoutes.POST("", wishlistHandler.AddToWishlist)
		wishlistRoutes.DELETE("/:game_id", wishlistHandler.RemoveFromWishlist)
	}

	router.POST("/checkout", auth.RequireAuth(tokenManager), checkoutHandler.Checkout)

	orderRoutes := router.Group("/orders", auth.RequireAuth(tokenManager))
	{
		orderRoutes.GET("", orderHandler.ListOrders)
		orderRoutes.GET("/:id", orderHandler.GetOrder)
	}

	router.GET("/library", auth.RequireAuth(tokenManager), libraryHandler.GetLibrary)

	router.GET("/games/:id/reviews", reviewHandler.ListReviews)
	router.POST("/games/:id/reviews", auth.RequireAuth(tokenManager), reviewHandler.CreateReview)
	router.PUT("/reviews/:id", auth.RequireAuth(tokenManager), reviewHandler.UpdateReview)
	router.DELETE("/reviews/:id", auth.RequireAuth(tokenManager), reviewHandler.DeleteReview)

	adminRoutes := router.Group("/admin", auth.RequireAuth(tokenManager), auth.AdminOnly(userRepo))
	{
		adminRoutes.POST("/games", gameHandler.CreateGame)
		adminRoutes.PUT("/games/:id", gameHandler.UpdateGame)
		adminRoutes.DELETE("/games/:id", gameHandler.DeleteGame)
	}

	log.Printf("starting reload api on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
