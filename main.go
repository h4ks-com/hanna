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
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	// Start HTTP API
	api := &API{bot: bot, token: apiToken}
	srv := &http.Server{Addr: apiAddr, Handler: api.routes()}

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

// --- HTTP API ---

type API struct {
	bot   *irc.Client
	token string
	mux   *http.ServeMux
}

type errorResponse struct{ Error string `json:"error"` }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (a *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.token == "" {
			writeJSON(w, http.StatusForbidden, errorResponse{"API_TOKEN not set on server"})
			return
		}
		auth := r.Header.Get("Authorization")
		const pfx = "Bearer "
		if !strings.HasPrefix(auth, pfx) || strings.TrimPrefix(auth, pfx) != a.token {
			writeJSON(w, http.StatusUnauthorized, errorResponse{"invalid or missing bearer token"})
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (a *API) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if a.bot.Connected() {
			writeJSON(w, 200, map[string]any{"ok": true, "nick": a.bot.Nick()})
		} else {
			writeJSON(w, 503, map[string]any{"ok": false})
		}
	})

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"version": Version, "name": "Hanna IRC Bot"})
	})

	mux.HandleFunc("/api/state", a.auth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"connected": a.bot.Connected(),
			"nick":      a.bot.Nick(),
			"channels":  a.bot.GetChannelStates(),
		})
	}))

	mux.HandleFunc("/api/join", a.auth(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Channel string `json:"channel"` }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Channel == "" {
			writeJSON(w, 400, errorResponse{"channel required"})
			return
		}
		a.bot.Join(in.Channel)
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}))

	mux.HandleFunc("/api/part", a.auth(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Channel, Reason string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Channel == "" {
			writeJSON(w, 400, errorResponse{"channel required"})
			return
		}
		a.bot.Part(in.Channel, in.Reason)
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}))

	mux.HandleFunc("/api/send", a.auth(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Target, Message string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Target == "" || in.Message == "" {
			writeJSON(w, 400, errorResponse{"target and message required"})
			return
		}
		a.bot.Privmsg(in.Target, in.Message)
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}))

	mux.HandleFunc("/api/nick", a.auth(func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Nick string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Nick) == "" {
			writeJSON(w, 400, errorResponse{"nick required"})
			return
		}
		a.bot.SetNick(in.Nick)
		writeJSON(w, 200, map[string]string{"status": "ok"})
	}))

	a.mux = mux
	return mux
}