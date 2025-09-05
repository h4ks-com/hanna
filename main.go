// main.go
// A self-contained Go IRC bot that connects over TLS and exposes a minimal
// token-authenticated HTTP API to control it.
//
// Features
// - TLS IRC connection (with optional server password)
// - Optional SASL PLAIN authentication
// - Auto-reconnect with exponential backoff
// - Graceful shutdown
// - Token-protected REST API endpoints for join/part/send/raw/nick/state
// - Simple channel tracking and PING/PONG handling
// - N8N webhook integration for chat processing
//
// Configuration: See .env.example for all environment variables
//
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hanna/irc"
)

const Version = "2.0.0"

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Hanna IRC Bot v%s starting up...", Version)

	apiToken := os.Getenv("API_TOKEN")
	apiAddr := getenv("API_ADDR", ":"+getenv("API_PORT", "8080"))
	apiTLS := boolenv("API_TLS", false)
	apiCert := os.Getenv("API_CERT")
	apiKey := os.Getenv("API_KEY")

	// Validate TLS configuration
	if apiTLS && (apiCert == "" || apiKey == "") {
		log.Fatalf("API_CERT and API_KEY are required when API_TLS=1")
	}

	bot := irc.NewClient()
	sup := NewSupervisor(bot)

	// Run IRC supervisor
	go sup.Run()

	// Start HTTP API using the comprehensive API from the IRC client
	srv := &http.Server{Addr: apiAddr, Handler: bot.CreateAPI(apiToken)}

	go func() {
		if apiTLS {
			log.Printf("HTTPS API listening on %s", apiAddr)
			if err := srv.ListenAndServeTLS(apiCert, apiKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("https server error: %v", err)
			}
		} else {
			log.Printf("HTTP API listening on %s", apiAddr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("http server error: %v", err)
			}
		}
	}()

	// Graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	log.Printf("shutting down...")

	sup.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)

	log.Printf("bye")
}

// Helper functions
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolenv(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "1" || v == "true"
	}
	return def
}

type Supervisor struct {
	client *irc.Client
	stop   chan struct{}
}

func NewSupervisor(c *irc.Client) *Supervisor {
	return &Supervisor{client: c, stop: make(chan struct{})}
}

func (s *Supervisor) Run() {
	backoff := time.Second
	max := 2 * time.Minute

	for {
		select {
		case <-s.stop:
			log.Printf("Supervisor stopping")
			return
		default:
		}

		log.Printf("Attempting to connect...")
		ctx := context.Background()
		if err := s.client.Dial(ctx); err != nil {
			log.Printf("dial error: %v", err)
		} else {
			log.Printf("Waiting for IRC registration...")
			time.Sleep(2 * time.Second)
		}

		// Wait until connection drops
		for s.client.Connected() {
			time.Sleep(500 * time.Millisecond)
		}

		// Backoff before reconnect
		log.Printf("disconnected; reconnecting in %s", backoff)
		select {
		case <-time.After(backoff):
		case <-s.stop:
			log.Printf("Supervisor stopping during backoff")
			return
		}
		backoff *= 2
		if backoff > max {
			backoff = max
		}
	}
}

func (s *Supervisor) Stop() {
	log.Printf("Stopping supervisor")
	close(s.stop)
	_ = s.client.Close()
}