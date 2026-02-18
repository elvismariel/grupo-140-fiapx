package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	inbound_messaging "video-processor-worker/internal/adapters/inbound/messaging"
	outbound_email "video-processor-worker/internal/adapters/outbound/email"
	outbound_processor "video-processor-worker/internal/adapters/outbound/processor"
	outbound_repository "video-processor-worker/internal/adapters/outbound/repository"
	outbound_storage "video-processor-worker/internal/adapters/outbound/storage"
	core_services "video-processor-worker/internal/core/services"

	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fmt.Println("🚀 Video Processor Worker starting...")

	// Start Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📊 Metrics server started on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Printf("⚠️ Metrics server failed: %v", err)
		}
	}()

	// Create root context with cancellation for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Verify dependencies
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("❌ Error: ffmpeg not found in system")
	}

	// Database initialization
	dbPool, err := initDatabase(ctx)
	if err != nil {
		log.Fatal("❌ Error initializing database: ", err)
	}
	defer dbPool.Close()

	// Initialize Adapters
	storage := outbound_storage.NewFSStorage()
	processor := outbound_processor.NewFFmpegProcessor()
	videoRepo := outbound_repository.NewPostgresVideoRepository(dbPool)
	userRepo := outbound_repository.NewPostgresUserRepository(dbPool)
	emailer := outbound_email.NewLogEmailAdapter()

	// Initialize Core Service
	worker := core_services.NewWorkerService(processor, storage, videoRepo, userRepo, emailer)

	// Initialize Inbound Adapters (NATS and Postgresql Poller)

	// 1. NATS Consumer
	natsURL := getEnv("NATS_URL", "nats://nats1:4222")
	consumer, err := inbound_messaging.NewNatsConsumerAdapter(natsURL, worker.ProcessVideoByID)
	if err != nil {
		log.Printf("⚠️ Error connecting to NATS: %v. Fallback to polling only.", err)
	} else {
		go func() {
			if err := consumer.Listen(ctx); err != nil {
				log.Printf("⚠️ NATS listener stopped: %v", err)
			}
		}()
	}

	// 2. Poller (Fallback Postgresql)
	//poller := polling.NewPollerAdapter(videoRepo, worker.ProcessVideoByID)
	//go poller.Start(ctx)

	log.Println("✅ Worker is up and running. Press Ctrl+C to stop.")

	// Wait for termination signal
	<-ctx.Done()
	log.Println("👋 Shutting down worker gracefully...")

	// Give some time for ongoing tasks to finish if needed
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("🛑 Worker stopped.")
}

func initDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := getEnv("DB_HOST", "db")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
		}
		log.Printf("⏳ Waiting for database... (%d/10)", i+1)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
	return nil, err
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
