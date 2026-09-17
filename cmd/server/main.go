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
	storerpc "github.com/wilson-lyc/dextea-store-service/internal/rpc"
)

func main() {
	defaultConfig := config.Default()
	addr := flag.String("addr", defaultConfig.Server.Addr, "gRPC 监听地址")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("[fatal] listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	// 业务 StoreService RPC 待协议仓库定义后注册到这里。
	_ = storerpc.NewServer()

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("[info] dextea-store-service gRPC listening on %s", *addr)
		if serveErr := grpcServer.Serve(listener); serveErr != nil {
			log.Printf("[error] gRPC server stopped: %v", serveErr)
		}
	}()

	<-ctx.Done()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

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
