package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/liftel/epic-fain/internal/application"
	"github.com/liftel/epic-fain/internal/domain/service"
	"github.com/liftel/epic-fain/internal/infrastructure/adapter/inbound/tcp"
	"github.com/liftel/epic-fain/internal/infrastructure/adapter/outbound/persistence"
	"github.com/liftel/epic-fain/internal/infrastructure/config"

	httpAdapter "github.com/liftel/epic-fain/internal/infrastructure/adapter/inbound/http"
)

func main() {
	cfg := config.Load()

	// --- Database ---
	db, err := persistence.NewPostgresDB(persistence.PostgresConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// --- Repositories ---
	telemetryRepo := persistence.NewTelemetryRepo(db)
	auditRepo := persistence.NewAuditLogRepo(db)
	_ = persistence.NewInstallationRepo(db)
	_ = persistence.NewAlertRepo(db)

	// --- Domain services ---
	decoder := service.NewCANDecoder()

	// --- Application services ---
	telemetrySvc := application.NewTelemetryAppService(
		telemetryRepo,
		auditRepo,
		nil, // alertSvc: will be wired in Hito 3
		nil, // mqttPub: will be wired in Hito 3
		decoder,
	)

	// --- Context with graceful shutdown ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// --- TCP Server (CAN frame receiver) ---
	tcpServer := tcp.NewServer(cfg.TCPAddr, telemetrySvc)
	go func() {
		if err := tcpServer.Start(ctx); err != nil {
			log.Printf("TCP server error: %v", err)
		}
	}()

	// --- HTTP API ---
	auth := httpAdapter.NewAPIKeyAuth(cfg.APIKeys)
	router := httpAdapter.NewRouter(
		telemetrySvc,
		nil, // controlSvc: requires CANSender adapter
		nil, // alertSvc: Hito 3
		nil, // installSvc: will be implemented
		auth,
	)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router.Handler(),
	}

	go func() {
		log.Printf("[HTTP] Listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// --- Wait for shutdown signal ---
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)
	cancel()

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
