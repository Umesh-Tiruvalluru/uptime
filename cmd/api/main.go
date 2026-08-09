package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/config"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/events"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/handler"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/repository"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/scheduler"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

var cfg config.Config

func main() {
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.NATSURL = os.Getenv("NATS_URL")
	cfg.Port = os.Getenv("PORT")
	cfg.JWTSecret = os.Getenv("JWT_SECRET")

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	dbpool := connectDB()
	defer dbpool.Close()
	nc := connectNATS()
	defer nc.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:          12 * time.Hour,
	}))

	repo := repository.NewRepository(dbpool)
	svc := service.NewService(repo, &cfg)
	h := handler.NewHandler(svc, &cfg)
	broker, err := events.NewBroker(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to subscribe to check results: %v\n", err)
		os.Exit(1)
	}

	schedulerCtx, stopScheduler := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopScheduler()
	go scheduler.Start(schedulerCtx, repo, nc)
	router.GET("/api/events", handler.SSE(broker))
	router.GET("/api/status", h.GetPublicStatus)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", h.RegisterUser)
			auth.POST("/login", h.LoginUser)
			auth.GET("/me", h.AuthMiddleware(), h.GetMe)
		}

		monitors := v1.Group("/monitors")
		monitors.Use(h.AuthMiddleware())
		{
			monitors.POST("/", h.CreateMonitor)
			monitors.GET("/", h.GetAllMonitors)
			monitors.GET("/:id", h.GetMonitorByID)
			monitors.PATCH("/:id", h.UpdateMonitor)
			monitors.DELETE("/:id", h.DeleteMonitor)
			monitors.GET("/:id/history", h.GetMonitorHistory)
		}
	}

	addr := ":" + cfg.Port
	if err := router.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server stopped: %v\n", err)
		os.Exit(1)
	}
}

func connectDB() *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	return dbpool
}

func connectNATS() *nats.Conn {
	if cfg.NATSURL == "" {
		cfg.NATSURL = nats.DefaultURL
	}
	nc, err := nats.Connect(cfg.NATSURL, nats.Name("monitoring-api"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to NATS: %v\n", err)
		os.Exit(1)
	}
	return nc
}
