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
	"google.golang.org/grpc/credentials"
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
	var grpcOpts []grpc.ServerOption
	grpcTLS := false
	switch {
	case cfg.GRPCTLSCertFile != "" && cfg.GRPCTLSKeyFile != "":
		creds, err := credentials.NewServerTLSFromFile(cfg.GRPCTLSCertFile, cfg.GRPCTLSKeyFile)
		if err != nil {
			log.Fatalf("grpc tls: %v", err)
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
		grpcTLS = true
	case cfg.GRPCTLSCertFile != "" || cfg.GRPCTLSKeyFile != "":
		log.Fatal("grpc tls: GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE must both be set")
	case cfg.Env == "production":
		log.Print("grpc tls: WARNING serving plaintext in production — set GRPC_TLS_CERT_FILE/GRPC_TLS_KEY_FILE or terminate TLS in a proxy")
	}
	g := grpc.NewServer(grpcOpts...)
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
		mode := "plaintext"
		if grpcTLS {
			mode = "tls"
		}
		log.Printf("gRPC listening on %s (%s)", cfg.GRPCAddr, mode)
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

	// GracefulStop alone would hang forever: agent bidi streams never end on
	// their own, and the held port makes a hub restart fail to bind.
	grpcDone := make(chan struct{})
	go func() {
		g.GracefulStop()
		close(grpcDone)
	}()
	select {
	case <-grpcDone:
	case <-time.After(5 * time.Second):
		g.Stop()
	}
	if err := h.Shutdown(ctx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	log.Println("stopped")
}
