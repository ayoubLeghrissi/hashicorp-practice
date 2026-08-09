package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"

	"github.com/services/pkg/logger"
)

// Config holds TLS file paths.
type Config struct {
	CertFile string
	KeyFile  string
	CAFile   string // Optional CA certificate for mutual TLS
}

// LoadServerCredentials loads TLS credentials for a gRPC server.
// Falls back to insecure mode if cert/key files are not found (for development).
func LoadServerCredentials(cfg Config, log *logger.Logger) (credentials.TransportCredentials, error) {
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		log.Warn("TLS cert/key files not configured — running WITHOUT TLS (development mode)")
		return nil, nil
	}

	// Check if files exist
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		log.Warn("TLS cert file not found at %s — running WITHOUT TLS (development mode)", cfg.CertFile)
		return nil, nil
	}
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		log.Warn("TLS key file not found at %s — running WITHOUT TLS (development mode)", cfg.KeyFile)
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	// Load CA cert for mutual TLS if provided
	if cfg.CAFile != "" {
		if _, err := os.Stat(cfg.CAFile); err == nil {
			ca, err := os.ReadFile(cfg.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(ca) {
				return nil, fmt.Errorf("failed to parse CA cert")
			}
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = pool
			log.Info("Mutual TLS enabled with CA: %s", cfg.CAFile)
		}
	}

	log.Info("TLS enabled with cert: %s", cfg.CertFile)
	return credentials.NewTLS(tlsConfig), nil
}

// LoadClientCredentials loads TLS credentials for a gRPC client.
// Falls back to insecure mode if cert file is not found.
func LoadClientCredentials(cfg Config, log *logger.Logger) (credentials.TransportCredentials, error) {
	if cfg.CertFile == "" {
		log.Warn("TLS cert not configured for client — using insecure connection")
		return nil, nil
	}

	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		log.Warn("TLS cert file not found at %s — using insecure connection", cfg.CertFile)
		return nil, nil
	}

	creds, err := credentials.NewClientTLSFromFile(cfg.CertFile, "")
	if err != nil {
		return nil, fmt.Errorf("failed to load client TLS credentials: %w", err)
	}

	log.Info("Client TLS enabled with cert: %s", cfg.CertFile)
	return creds, nil
}
