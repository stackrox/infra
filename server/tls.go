package server

import (
	"crypto/tls"
	"errors"
	"net"

	"github.com/stackrox/infra/config"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TLSManager represents a type that facilitates configuring a http server with
// TLS certificates.
type TLSManager interface {
	DialOptions() []grpc.DialOption
	Listener() net.Listener
	Name() string
	ServerOption() grpc.ServerOption
	TLSConfig() *tls.Config
}

var (
	_ TLSManager = (*letsEncryptManager)(nil)
	_ TLSManager = (*localCertManager)(nil)
)

type letsEncryptManager struct {
	*autocert.Manager
	domain string
}

func newLetsEncryptManager(domain, certDir string) TLSManager {
	return &letsEncryptManager{
		Manager: &autocert.Manager{
			Cache:      autocert.DirCache(certDir),
			HostPolicy: autocert.HostWhitelist(domain),
			Prompt:     autocert.AcceptTOS,
		},
		domain: domain,
	}
}

func (m letsEncryptManager) DialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, m.domain)),
	}
}

func (m letsEncryptManager) Name() string {
	return "LetsEncrypt"
}

func (m letsEncryptManager) ServerOption() grpc.ServerOption {
	return grpc.Creds(credentials.NewTLS(m.TLSConfig()))
}

type localCertManager struct {
	tlsConfig *tls.Config
	listener  net.Listener
}

func newLocalCertManager(certFile, keyFile string, httpsPort string) (TLSManager, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:"+httpsPort)
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return localCertManager{
		tlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"},
		},
		listener: listener,
	}, nil
}

func (m localCertManager) DialOptions() []grpc.DialOption {
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(cfg)),
	}
}

func (m localCertManager) Listener() net.Listener {
	return tls.NewListener(m.listener, m.TLSConfig())
}

func (m localCertManager) Name() string {
	return "local certificate+key"
}

func (m localCertManager) ServerOption() grpc.ServerOption {
	return grpc.Creds(credentials.NewTLS(m.TLSConfig()))
}

func (m localCertManager) TLSConfig() *tls.Config {
	return m.tlsConfig
}

// NewTLSManager creates an appropriate TLSManager based on the current server
// configuration.
func NewTLSManager(serverCfg config.ServerConfig) (TLSManager, error) {
	switch {
	case serverCfg.CertFile != "" && serverCfg.KeyFile != "":
		return newLocalCertManager(serverCfg.CertFile, serverCfg.KeyFile, serverCfg.HTTPS)

	case serverCfg.HTTPS == "443" && serverCfg.Domain != "":
		return newLetsEncryptManager(serverCfg.Domain, serverCfg.CertDir), nil

	default:
		return nil, errors.New("invalid server configuration")
	}
}
