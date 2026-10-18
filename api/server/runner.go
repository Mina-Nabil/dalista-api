package server

import (
	"fmt"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/api/server/middlewares"
	"dalista/dalista-be/api/server/routes"
	"dalista/dalista-be/config"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Run builds the Gin engine, attaches middleware, and serves on :8080.
func Run() {
	app := gin.Default()
	RegisterCustomValidators()

	cfg := config.GetConfig()

	// Trust only the VPC CIDR so c.ClientIP() honors X-Forwarded-For from the
	// ALB (which sits inside the VPC) but rejects spoofed values from anywhere
	// else. Gin's default of trusting all proxies would let any client forge the
	// source IP and bypass per-IP rate limiting.
	if err := app.SetTrustedProxies([]string{"10.0.0.0/16"}); err != nil {
		logger.Fatal(err, "Failed to configure trusted proxies")
	}

	fmt.Println("CORS Allowed Origins: ", cfg.CorsAllowedOrigins())

	corsConfig := cors.New(cors.Config{
		AllowOrigins:     cfg.CorsAllowedOrigins(),
		AllowCredentials: true,
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})
	app.Use(corsConfig)

	// Attach the Sentry middleware.
	app.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	// Swagger documentation — development only. Exposing the full API surface in
	// staging/production hands reconnaissance scanners a complete endpoint map.
	if cfg.AppEnv == "development" {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ALB health check. Registered before the rate limiter so load-balancer
	// probes are never throttled. The versioned GET /api/v1/health lives in the
	// route table (see routes/api.go).
	app.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Per-IP rate limiter (token bucket). Tunable via RATE_LIMIT_PER_MINUTE; 0 disables.
	app.Use(middlewares.RateLimitMiddleware(cfg.RateLimitPerMinute))

	routes.LoadRoutes(app)

	app.Run(":8080")
}
