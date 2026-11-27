package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/erayarslan/go-bladerf/mock"
)

func main() {
	listenAddr := flag.String("l", "0.0.0.0:5555", "TCP listen address (host:port)")
	flag.Parse()

	srv, err := mock.NewServer(*listenAddr)
	if err != nil {
		log.Fatalf("failed to start bladerf-mock: %v", err)
	}

	log.Printf("bladerf-mock listening on %s (simulated bladeRF 2.0 xA4)", srv.Addr())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down bladerf-mock")
	if err := srv.Close(); err != nil {
		log.Fatalf("close error: %v", err)
	}
}
