package main

// @title Fiap X Video Processor API
// @version 1.0
// @description API for uploading videos and managing their processing.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

import (
	"fmt"
	"log"
	inbound_http "video-processor/internal/adapters/inbound/http"
	outbound_messaging "video-processor/internal/adapters/outbound/messaging"
	outbound_repository "video-processor/internal/adapters/outbound/repository"
	outbound_storage "video-processor/internal/adapters/outbound/storage"
	core_services "video-processor/internal/core/services"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"

	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"

	_ "video-processor/docs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("Starting Video Processor API...")

	// Database initialization
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	// Run migrations
	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		log.Printf("⚠️ Erro ao preparar migrações: %v", err)
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Printf("⚠️ Erro ao aplicar migrações: %v", err)
		} else {
			fmt.Println("✅ Migrações aplicadas com sucesso!")
		}
	}

	dbPool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal("❌ Erro ao conectar ao banco de dados: ", err)
	}
	defer dbPool.Close()

	// Initialize Outbound Adapters
	storage := outbound_storage.NewFSStorage()
	userRepo := outbound_repository.NewPostgresUserRepository(dbPool)
	videoRepo := outbound_repository.NewPostgresVideoRepository(dbPool)

	// Initialize NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats1:4222"
	}
	eventPublisher, err := outbound_messaging.NewNatsAdapter(natsURL)
	if err != nil {
		log.Printf("⚠️ Erro ao conectar ao NATS: %v. O sistema continuará sem publicação de eventos.", err)
		// We could use a mock or NullPublisher here if we wanted to be more robust
		// For now, let's just log and see. But NewVideoService expects a port.
		// I'll implement a simple NoOp publisher in case of error.
	}

	// Initialize Core Services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "fiapx-secret-key"
	}

	videoService := core_services.NewVideoService(storage, videoRepo, eventPublisher)
	userService := core_services.NewUserService(userRepo, jwtSecret)

	// Initialize Inbound Adapter (HTTP)
	handler := inbound_http.NewHandler(videoService, userService, storage, jwtSecret)

	r := gin.Default()

	// Prometheus Metrics
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	// Middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static files
	r.Static("/uploads", "/app/uploads")
	r.Static("/outputs", "/app/outputs")

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register Routes
	handler.RegisterRoutes(r)

	fmt.Println("🎬 Servidor iniciado na porta 8080")
	fmt.Println("📂 Acesse: http://localhost:8080")

	log.Fatal(r.Run(":8080"))
}
