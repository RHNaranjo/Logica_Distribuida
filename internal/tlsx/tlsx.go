package tlsx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc/credentials"

	"logica/internal/config"
)

// Leer el certificado propio desde la carpeta
func cargar() (tls.Certificate, *x509.CertPool, error) {
	dir := config.DirectorioCerts("/etc/ssl/certs/grpc")

	propio, err := tls.LoadX509KeyPair(filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key"))
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("no se puede leer el certificado propio: %w", err)
	}

	// Leer el certificado crudo
	crudo, err := os.ReadFile(filepath.Join(dir, "ca.crt"))
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("no se pudo leer el certificado de autoridad %w", err)
	}

	// Parsear el certificado
	ca := x509.NewCertPool()
	if !ca.AppendCertsFromPEM(crudo) {
		return tls.Certificate{}, nil, fmt.Errorf("ca.crt no trae ningún certificado válido")
	}

	return propio, ca, nil
}

// Exigir que el cliente tenga un certificado firmado
func Servidor() (credentials.TransportCredentials, error) {
	propio, ca, err := cargar()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{propio},      // Certificado digital y llave privada
		ClientAuth:   tls.RequireAndVerifyClientCert, // Modo mTLS estricto
		ClientCAs:    ca,                             // Autoridad certificadora
		MinVersion:   tls.VersionTLS13,               // Sólo TLS 1.3
	}), nil
}

// Presentar el certificado propio y verificar
func Cliente(servidor string) (credentials.TransportCredentials, error) {
	propio, ca, err := cargar()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{propio},
		RootCAs:      ca,
		ServerName:   servidor,
		MinVersion:   tls.VersionTLS13,
	}), nil
}
