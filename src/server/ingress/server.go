package ingress

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	certificates2 "server/certificates"
	"server/tools"
	"strconv"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

type ServerListener struct {
	SetupHandler           *SetupHandler
	ServerCertificateCache certificates2.CertificateCache
	Config                 *tools.GlobalConfig
	CertificateService     certificates2.CertificateService
}

func (s *ServerListener) OpenPortsAndListen() ([]*http.Server, error) {
	authHandler := http.HandlerFunc(s.SetupHandler.ProxyMiddleware)
	redirectingMiddleWare := RedirectingMiddleware{}
	handler := u.RecoverHttpPanics(redirectingMiddleWare.Handler(authHandler))

	httpServer := &http.Server{
		Addr:              ":80",
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       1 * time.Minute,
	}

	tlsServer, err := s.buildTlsServer(handler)
	if err != nil {
		return nil, err
	}

	go s.runServer(func() error { return httpServer.ListenAndServe() })
	go s.runServer(func() error { return tlsServer.ListenAndServeTLS("", "") })

	return []*http.Server{httpServer, tlsServer}, nil
}

func (s *ServerListener) buildTlsServer(handler http.Handler) (*http.Server, error) {
	bundle, err := s.CertificateService.GetCurrentCertificate()
	if err != nil {
		return nil, err
	}
	s.ServerCertificateCache.SetCertificate(bundle.GetTlsCertificate())

	return &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			MinVersion:     tls.VersionTLS12,
			GetCertificate: s.dynamicCertProvider(),
		},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          log.New(errorLoggerWhichDropsBadCertMessages{}, "", 0),
	}, nil
}

func (s *ServerListener) runServer(listen func() error) {
	if err := listen(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		u.Logger.Error(err)
	}
}

type errorLoggerWhichDropsBadCertMessages struct{}

func (w errorLoggerWhichDropsBadCertMessages) Write(p []byte) (int, error) {
	if isNetworkChangedError(p) {
		return len(p), nil
	}

	rawMessage := strings.TrimSpace(string(p))
	unescapedMessage := rawMessage
	if unquotedMessage, err := strconv.Unquote(rawMessage); err == nil {
		unescapedMessage = unquotedMessage
	}

	for line := range strings.SplitSeq(unescapedMessage, "\n") {
		trimmedLine := strings.TrimRight(line, "\r\t ")
		if trimmedLine != "" {
			u.Logger.Error(trimmedLine)
		}
	}

	return len(p), nil
}

func isNetworkChangedError(payload []byte) bool {
	// "visit https://quollix.org and search for 'NETWORK_CHANGED' for more information"
	payloadString := string(payload)
	return strings.Contains(payloadString, "TLS handshake error") && strings.Contains(payloadString, "unknown certificate")
}

func (s *ServerListener) dynamicCertProvider() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert := s.ServerCertificateCache.GetCertificate()
		if cert == nil {
			return nil, u.Logger.NewError("no certificate found")
		}
		return cert, nil
	}
}
