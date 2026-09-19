package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/wilson-lyc/dextea-store-service/internal/config"
	"github.com/wilson-lyc/dextea-store-service/internal/registry"
	"github.com/wilson-lyc/dextea-store-service/internal/repository"
	storerpc "github.com/wilson-lyc/dextea-store-service/internal/rpc"
	"github.com/wilson-lyc/dextea-store-service/internal/service"
	storev1 "github.com/wilson-lyc/dextea-v3-proto/gen/go/store/v1"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("[fatal] %v", err)
	}

	db, err := repository.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("[fatal] %v", err)
	}
	defer db.Close()

	storeService := service.NewStoreService(repository.NewStoreRepository(db))

	listener, err := net.Listen("tcp", cfg.Server.Addr)
	if err != nil {
		log.Fatalf("[fatal] listen: %v", err)
	}

	serverOptions := make([]grpc.ServerOption, 0, 1)
	if cfg.Auth.Enabled {
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(storerpc.RoleTokenInterceptor(storerpc.RoleTokens{
			Admin:      cfg.Auth.AdminToken,
			Business:   cfg.Auth.BusinessToken,
			Credential: cfg.Auth.CredentialToken,
		})))
	}
	grpcServer := grpc.NewServer(serverOptions...)
	storev1.RegisterStoreAdminServiceServer(grpcServer, storerpc.NewAdminServer(storeService))
	storev1.RegisterStoreBusinessServiceServer(grpcServer, storerpc.NewBusinessServer(storeService))
	storev1.RegisterStoreCredentialServiceServer(grpcServer, storerpc.NewCredentialServer(storeService))

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("[info] dextea-store-service gRPC listening on %s", cfg.Server.Addr)
		if serveErr := grpcServer.Serve(listener); serveErr != nil {
			log.Printf("[error] gRPC server stopped: %v", serveErr)
		}
	}()
	var reg *registry.Registrar
	if cfg.Nacos.Enabled {
		reg, err = registry.Register(cfg.Nacos, cfg.Server.Addr)
		if err != nil {
			log.Fatalf("[fatal] %v", err)
		}
	}

	<-ctx.Done()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	if reg != nil {
		reg.Deregister()
	}

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		grpcServer.Stop()
	}

	log.Println("[info] dextea-store-service stopped")
}
