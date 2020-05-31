package main

import (
	"1v1.group/zlog/webhook"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Cfg struct {
	port     int
	certFile string
	keyFile  string
}

func main() {
	var cfg Cfg

	flag.IntVar(&cfg.port, "port", 443, "Webhook server port.")
	flag.StringVar(&cfg.certFile, "tls-cert-file", "/etc/pki/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&cfg.keyFile, "tls-key-file", "/etc/pki/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	pair, err := tls.LoadX509KeyPair(cfg.certFile, cfg.keyFile)

	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhook.Serve)

	server := http.Server{
		Addr:      fmt.Sprintf(":%v", cfg.port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		Handler:   mux,
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	log.Println("Server started")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Got OS shutdown signal, shutting down webhook server gracefully...")
	_ = server.Shutdown(context.Background())
}
