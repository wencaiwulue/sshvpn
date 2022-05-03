package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"embed"

	log "github.com/sirupsen/logrus"
)

//go:embed server.crt
var crt embed.FS

//go:embed server.key
var key embed.FS

var TlsConfigServer *tls.Config
var TlsConfigClient *tls.Config

func init() {
	crtBytes, _ := crt.ReadFile("server.crt")
	keyBytes, _ := key.ReadFile("server.key")
	pair, err := tls.X509KeyPair(crtBytes, keyBytes)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(crtBytes)

	TlsConfigServer = &tls.Config{
		Certificates: []tls.Certificate{pair},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}

	TlsConfigClient = &tls.Config{
		Certificates:       []tls.Certificate{pair},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
		ClientSessionCache: tls.NewLRUClientSessionCache(1 << 10),
	}
}
