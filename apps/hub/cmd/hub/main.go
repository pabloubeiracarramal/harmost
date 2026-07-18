package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/harmost/hub/internal/events"
	"github.com/harmost/hub/internal/platform"
	"github.com/harmost/hub/internal/service"
	"github.com/harmost/hub/internal/transport/grpcapi"
	"github.com/harmost/hub/internal/transport/httpapi"
	"google.golang.org/grpc"
)

func main() {
	cfg := platform.LoadConfig()

	db, err := platform.NewDB(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	bus := events.New()
	svc := service.New(db, cfg.FrontendURL, bus)
	grpcSrv := grpcapi.New(svc, bus)
	httpSrv := httpapi.New(svc, grpcSrv, bus, cfg)

	// No streams exist yet — flip agents left online by a previous crash so
	// the orphan sweeper can see their jobs.
	if err := svc.Agent.MarkAllOffline(context.Background()); err != nil {
		log.Printf("mark agents offline: %v", err)
	}

	// Orphan sweeper: fail jobs whose agent has been offline > 2m.
	sweepCtx, stopSweeper := context.WithCancel(context.Background())
	defer stopSweeper()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-sweepCtx.Done():
				return
			case <-ticker.C:
				if err := svc.Job.SweepOrphans(sweepCtx, 2*time.Minute); err != nil {
					log.Printf("sweep orphans: %v", err)
				}
			}
		}
	}()

	// ── gRPC ────────────────────────────────────────────────────────────────
	g := grpc.NewServer()
	grpcSrv.Register(g)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	// ── HTTP ─────────────────────────────────────────────────────────────────
	h := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      httpSrv.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── start ────────────────────────────────────────────────────────────────
	go func() {
		log.Printf("gRPC listening on %s", cfg.GRPCAddr)
		if err := g.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	go func() {
		log.Printf("HTTP listening on %s", cfg.HTTPAddr)
		if err := h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	// ── graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	g.GracefulStop()
	if err := h.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	log.Println("stopped")
}
