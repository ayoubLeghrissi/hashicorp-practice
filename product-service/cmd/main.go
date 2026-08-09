package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/services/pkg/config"
	"github.com/services/pkg/db"
	"github.com/services/pkg/logger"
	tlspkg "github.com/services/pkg/tls"
	pb "github.com/services/proto/product"
	"github.com/services/product-service/repository"
	"github.com/services/product-service/service"
)

func main() {
	// Logger
	log, err := logger.New(logger.Config{
		ServiceName: "product-service",
		LogDir:      config.GetEnv("LOG_DIR", "./logs"),
		LogFileName: "access.log",
		Level:       logger.INFO,
		ToStdout:    true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting Product Service...")

	// Database
	dbCfg := db.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5432),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   config.GetEnv("DB_NAME", "product_db"),
		SSLMode:  config.GetEnv("DB_SSLMODE", "disable"),
	}

	database, err := db.Connect(dbCfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	repo := repository.NewProductRepository(database, log)
	if err := repo.Migrate(); err != nil {
		log.Fatal("Failed to run migrations: %v", err)
	}

	// TLS
	tlsCfg := tlspkg.Config{
		CertFile: config.GetEnv("TLS_CERT_FILE", "./certs/server.crt"),
		KeyFile:  config.GetEnv("TLS_KEY_FILE", "./certs/server.key"),
		CAFile:   config.GetEnv("TLS_CA_FILE", ""),
	}

	creds, err := tlspkg.LoadServerCredentials(tlsCfg, log)
	if err != nil {
		log.Fatal("Failed to load TLS credentials: %v", err)
	}

	// gRPC server
	var grpcServer *grpc.Server
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor(log)),
	}
	if creds != nil {
		serverOpts = append(serverOpts, grpc.Creds(creds))
	}
	grpcServer = grpc.NewServer(serverOpts...)

	// Register service
	productService := service.NewProductService(repo, log)
	pb.RegisterProductServiceServer(grpcServer, productService)

	// Enable reflection for development tools
	reflection.Register(grpcServer)

	// Listen
	port := config.GetEnv("GRPC_PORT", "50051")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen on port %s: %v", port, err)
	}

	log.Info("Product Service listening on :%s (TLS=%v)", port, creds != nil)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("Shutting down Product Service...")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("gRPC server failed: %v", err)
	}
}
