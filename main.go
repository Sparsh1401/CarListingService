package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/car-listing-service/config"
	"github.com/yourusername/car-listing-service/controllers"
	"github.com/yourusername/car-listing-service/database"
	"github.com/yourusername/car-listing-service/middleware"
	"github.com/yourusername/car-listing-service/repository"
	"github.com/yourusername/car-listing-service/routes"
	"github.com/yourusername/car-listing-service/services"
	"github.com/gin-gonic/gin"
)

func setupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "car-listing-service",
		})
	})

	carRepo := repository.NewCarRepository(database.DB)
	carService := services.NewCarService(carRepo)
	carController := controllers.NewCarController(carService)

	routes.SetupRoutes(router, carController)

	return router
}

func main() {
	cfg := config.LoadConfig()

	database.InitDB()

	router := setupRouter(cfg)

	srv := &http.Server{
		Addr:           ":" + cfg.ServerPort,
		Handler:        router,
		ReadTimeout:    300 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("Starting server on port %s in %s mode", cfg.ServerPort, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
