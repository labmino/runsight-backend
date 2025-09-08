package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/labmino/runsight-backend/internal/database"
	"github.com/labmino/runsight-backend/internal/handlers"
	"github.com/labmino/runsight-backend/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	authHandler := handlers.NewAuthHandler(db)
	mobileHandler := handlers.NewMobileHandler(db)
	iotHandler := handlers.NewIoTHandler(db)

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "RunSight API is running",
			})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware())
		{
			protectedAuth.GET("/profile", authHandler.GetProfile)
			protectedAuth.PUT("/profile", authHandler.UpdateProfile)
		}

		mobile := api.Group("/mobile")
		mobile.Use(middleware.AuthMiddleware())
		{
			mobile.POST("/pairing/request", mobileHandler.RequestPairingCode)
			mobile.GET("/pairing/:session_id/status", mobileHandler.CheckPairingStatus)

			mobile.GET("/devices", mobileHandler.GetDevices)
			mobile.DELETE("/devices/:device_id", mobileHandler.RemoveDevice)

			mobile.GET("/runs", mobileHandler.ListRuns)
			mobile.GET("/runs/:run_id", mobileHandler.GetRun)
			mobile.PATCH("/runs/:run_id", mobileHandler.UpdateRunNotes)
			mobile.GET("/stats", mobileHandler.GetStats)
		}

		iot := api.Group("/iot")
		{
			iot.POST("/pairing/verify", iotHandler.VerifyPairingCode)

			iotProtected := iot.Group("")
			iotProtected.Use(middleware.DeviceAuthMiddleware(db))
			{
				iotProtected.POST("/runs/upload", iotHandler.UploadRun)
				iotProtected.POST("/runs/batch", iotHandler.BatchUploadRuns)

				iotProtected.POST("/devices/status", iotHandler.UpdateDeviceStatus)
				iotProtected.GET("/devices/config", iotHandler.GetDeviceConfig)
			}
		}
	}


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}