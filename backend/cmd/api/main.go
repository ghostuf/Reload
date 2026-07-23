package main

import (
	"log"

	"github.com/ghostuf/reload/internal/config"
	"github.com/ghostuf/reload/internal/handlers"
	"github.com/ghostuf/reload/internal/repository"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	cfg := config.Load()

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	gameRepo := repository.NewGameRepository(db)
	gameService := services.NewGameService(gameRepo)
	gameHandler := handlers.NewGameHandler(gameService)

	router := gin.Default()
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/games", gameHandler.ListGames)
	router.GET("/games/:id", gameHandler.GetGame)

	log.Printf("starting reload api on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
