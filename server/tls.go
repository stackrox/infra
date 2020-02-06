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

func (m localCertManager) DialOptions() []grpc.DialOption {
	cfg := &tls.Config{
		InsecureSkipVerify: true, // nolint:gosec
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

func NewTLSManager(serverCfg config.ServerConfig) (TLSManager, error) {
	switch {
	case serverCfg.CertFile != "" && serverCfg.KeyFile != "":
		listener, err := net.Listen("tcp", "0.0.0.0:"+serverCfg.HTTPS)
		if err != nil {
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(serverCfg.CertFile, serverCfg.KeyFile)
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

	case serverCfg.HTTPS == "443" && serverCfg.Domain != "":
		return &letsEncryptManager{
			Manager: &autocert.Manager{
				Cache:      autocert.DirCache(serverCfg.CertDir),
				HostPolicy: autocert.HostWhitelist(serverCfg.Domain),
				Prompt:     autocert.AcceptTOS,
			},
			domain: serverCfg.Domain,
		}, nil

	default:
		return nil, errors.New("invalid server config")
	}
}
