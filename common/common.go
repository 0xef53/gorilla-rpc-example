package common

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

type ServerSummary struct {
	ServerName string
	Qemu       struct {
		Major int
		Minor int
		Micro int
	}
}

type ServerSummaryQuery struct {
	ServerName string
}

func TlsConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	if len(cert.Certificate) != 2 {
		return nil, fmt.Errorf("certificate should have 2 concatenated certificates: server + CA")
	}

	ca, err := x509.ParseCertificate(cert.Certificate[1])
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(ca)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      certPool,
		ClientCAs:    certPool,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		ClientSessionCache:       tls.NewLRUClientSessionCache(0),
	}, nil
}
