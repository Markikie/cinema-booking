package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Markikie/cinema-booking/internal/config"
	"github.com/Markikie/cinema-booking/internal/database"
	"github.com/Markikie/cinema-booking/internal/handlers"
	"github.com/Markikie/cinema-booking/internal/middleware"
	"github.com/Markikie/cinema-booking/internal/models"
	"github.com/Markikie/cinema-booking/internal/repository"
	"github.com/Markikie/cinema-booking/internal/service"
	"github.com/Markikie/cinema-booking/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ===== config =====
	cfg := config.Load()

	// ===== database =====
	mongoDB := database.NewMongoClient(cfg.MongoURI, cfg.MongoDBName)
	database.EnsureIndexes(mongoDB)

	redisClient := database.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)

	if err := service.EnableKeyspaceNotifications(context.Background(), redisClient); err != nil {
		log.Fatalf("failed to enable redis keyspace notifications: %v", err)
	}

	// ===== repository layer =====
	seatRepo := repository.NewSeatRepository(mongoDB)
	bookingRepo := repository.NewBookingRepository(mongoDB)
	auditRepo := repository.NewAuditLogRepository(mongoDB)
	userRepo := repository.NewUserRepository(mongoDB)

	// ===== WebSocket hub + pub/sub =====
	hub := ws.NewHub()
	pubsub := service.NewPubSubService(redisClient, hub)
	pubsub.StartSubscriber(context.Background()) // เริ่มฟัง redis pub/sub แบบ background goroutine

	// ===== service layer (business logic) =====
	lockService := service.NewLockService(redisClient, cfg.SeatLockTTL)
	bookingService := service.NewBookingService(seatRepo, bookingRepo, auditRepo, lockService, pubsub, cfg.SeatLockTTL)

	// expiry listener
	expiryListener := service.NewExpiryListener(redisClient, bookingService)
	expiryListener.Start(context.Background())

	// ===== handler layer (HTTP endpoints) =====
	authHandler := handlers.NewAuthHandler(userRepo, cfg.GoogleClientID, cfg.JWTSecret)
	bookingHandler := handlers.NewBookingHandler(bookingService, seatRepo)
	adminHandler := handlers.NewAdminHandler(bookingRepo, auditRepo)

	// ===== Gin router =====
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ----- public routes -----
	router.POST("/api/auth/login", authHandler.Login)
	router.GET("/ws", ws.ServeWS(hub))

	// ----- authenticated routes -----
	authorized := router.Group("/api")
	authorized.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		authorized.GET("/showtimes/:showtime_id/seats", bookingHandler.GetSeatMap)
		authorized.POST("/bookings/select-seat", bookingHandler.SelectSeat)
		authorized.POST("/bookings", bookingHandler.CreateBooking)
		authorized.POST("/bookings/confirm-payment", bookingHandler.ConfirmPayment)
	}

	// ----- admin-only routes -----
	admin := router.Group("/api/admin")
	admin.Use(middleware.RequireAuth(cfg.JWTSecret), middleware.RequireRole(models.RoleAdmin))
	{
		admin.GET("/bookings", adminHandler.ListBookings)
		admin.GET("/audit-logs", adminHandler.ListAuditLogs)
	}

	// ===== HTTP server =====
	log.Printf("server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
