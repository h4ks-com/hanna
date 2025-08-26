// main.go
// A self-contained Go IRC bot that connects over TLS and exposes a minimal
// token-authenticated HTTP API to control it.
//
// Features
// - TLS IRC connection (with optional server password)
// - Optional SASL PLAIN authentication (set SASL_USER/SASL_PASS)
// - Auto-reconnect with exponential backoff
// - Graceful shutdown
// - Token-protected REST API endpoints for join/part/send/raw/nick/state
// - Simple channel tracking and PING/PONG handling
//
// Configuration via environment variables:
//   IRC_ADDR       : host:port (e.g. "irc.libera.chat:6697")
//   IRC_TLS        : "1" to enable TLS (default 1)
//   IRC_TLS_INSECURE: "1" to skip TLS verification (default 0)
//   IRC_PASS       : optional server password
//   IRC_NICK       : bot nick (default "goircbot")
//   IRC_USER       : username/ident (default "goircbot")
//   IRC_NAME       : realname/gecos (default "Go IRC Bot")
//   SASL_USER      : SASL PLAIN authcid (optional)
//   SASL_PASS      : SASL PLAIN password (optional)
//   API_ADDR       : HTTP listen address (default ":8080")
//   API_TOKEN      : Bearer token required for API requests
//   API_TLS        : "1" to enable HTTPS (default 0)
//   API_CERT       : Path to TLS certificate file (required when API_TLS=1)
//   API_KEY        : Path to TLS private key file (required when API_TLS=1)
//   N8N_WEBHOOK    : n8n webhook URL for chat node integration (optional)
//   AUTOJOIN       : comma-separated channels to autojoin on connect (e.g. "#go,#bots")
//
// Build & Run
//   go mod init ircbot && go mod tidy
//   go build -o ircbot
//   API_TOKEN=secret IRC_ADDR=irc.example.net:6697 IRC_NICK=MyBot ./ircbot
//
// Example curl
//   curl -H 'Authorization: Bearer secret' -H 'Content-Type: application/json' \
//        -d '{"target":"#test","message":"hello from API"}' https://127.0.0.1:8080/api/send
//
package main

import (
    "bufio"
    "bytes"
    "context"
    "crypto/tls"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

// --- Utilities ---

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func boolenv(key string, def bool) bool {
    v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
    if v == "1" || v == "true" || v == "yes" {
        return true
    }
    if v == "0" || v == "false" || v == "no" {
        return false
    }
    return def
}

// --- IRC Client ---

type IRCClient struct {
    addr          string
    useTLS        bool
    tlsInsecure   bool
    pass          string
    nick          atomic.Value // string
    user          string
    name          string
    saslUser      string
    saslPass      string
    n8nWebhook    string

    conn   net.Conn
    rw     *bufio.ReadWriter
    wmu    sync.Mutex
    alive  atomic.Bool

    channelsMu sync.RWMutex
    channels   map[string]struct{}

    onReady func()
}

func NewIRCClient() *IRCClient {
    c := &IRCClient{
        addr:        getenv("IRC_ADDR", ""),
        useTLS:      boolenv("IRC_TLS", true),
        tlsInsecure: boolenv("IRC_TLS_INSECURE", false),
        pass:        os.Getenv("IRC_PASS"),
        user:        getenv("IRC_USER", "goircbot"),
        name:        getenv("IRC_NAME", "Go IRC Bot"),
        saslUser:    os.Getenv("SASL_USER"),
        saslPass:    os.Getenv("SASL_PASS"),
        n8nWebhook:  os.Getenv("N8N_WEBHOOK"),
        channels:    make(map[string]struct{}),
    }
    c.nick.Store(getenv("IRC_NICK", "goircbot"))
    return c
}

func (c *IRCClient) Connected() bool { return c.alive.Load() }

func (c *IRCClient) Nick() string { return c.nick.Load().(string) }

func (c *IRCClient) setNick(n string) { c.nick.Store(n) }

func (c *IRCClient) Dial(ctx context.Context) error {
    if c.addr == "" {
        return errors.New("IRC_ADDR is required")
    }
    log.Printf("Connecting to IRC server %s (TLS: %v)", c.addr, c.useTLS)
    var d net.Conn
    var err error
    if c.useTLS {
        tlsCfg := &tls.Config{InsecureSkipVerify: c.tlsInsecure}
        d, err = tls.Dial("tcp", c.addr, tlsCfg)
    } else {
        d, err = net.Dial("tcp", c.addr)
    }
    if err != nil {
        log.Printf("Connection failed: %v", err)
        return err
    }
    log.Printf("TCP connection established")
    c.conn = d
    c.rw = bufio.NewReadWriter(bufio.NewReader(d), bufio.NewWriter(d))

    // Registration sequence
    log.Printf("Starting IRC registration as nick: %s", c.Nick())
    if c.pass != "" {
        log.Printf("Sending server password")
        c.rawf("PASS %s", c.pass)
    }

    // Request SASL if configured
    sasl := c.saslUser != "" && c.saslPass != ""
    if sasl {
        log.Printf("Requesting SASL authentication")
        c.raw("CAP LS 302")
        c.raw("CAP REQ :sasl")
    }

    c.rawf("NICK %s", c.Nick())
    c.rawf("USER %s 0 * :%s", c.user, c.name)

    go c.readLoop()

    if sasl {
        // The rest of SASL handshake continues in readLoop upon CAP ACK
    }

    return nil
}

func (c *IRCClient) readLoop() {
    log.Printf("Starting IRC read loop")
    for {
        line, err := c.rw.ReadString('\n')
        if err != nil {
            log.Printf("IRC read error: %v", err)
            c.alive.Store(false)
            return
        }
        line = strings.TrimRight(line, "\r\n")
        if line == "" {
            continue
        }
        log.Printf("<< %s", line)
        c.handleLine(line)
    }
}

func (c *IRCClient) handleLine(line string) {
    prefix := ""
    rest := line
    if strings.HasPrefix(rest, ":") {
        if i := strings.Index(rest, " "); i != -1 {
            prefix = rest[1:i]
            rest = strings.TrimSpace(rest[i+1:])
        }
    }
    parts := strings.Split(rest, " ")
    if len(parts) == 0 {
        return
    }
    cmd := strings.ToUpper(parts[0])

    argEnd := strings.Index(rest, " :")
    var args []string
    var trailing string
    if argEnd != -1 {
        argsStart := len(cmd) + 1
        if argsStart < argEnd {
            args = strings.Fields(rest[argsStart:argEnd])
        }
        trailing = rest[argEnd+2:]
    } else {
        if len(rest) > len(cmd) {
            args = strings.Fields(rest[len(cmd)+1:])
        }
    }

    switch cmd {
    case "PING":
        if trailing == "" && len(args) > 0 {
            trailing = args[len(args)-1]
        }
        c.rawf("PONG :%s", trailing)
    case "001": // welcome
        log.Printf("IRC registration successful! Welcome message received")
        c.alive.Store(true)
        if c.onReady != nil {
            c.onReady()
        }
        // Autojoin
        if aj := strings.TrimSpace(os.Getenv("AUTOJOIN")); aj != "" {
            log.Printf("Auto-joining channels: %s", aj)
            for _, ch := range strings.Split(aj, ",") {
                ch = strings.TrimSpace(ch)
                if ch != "" {
                    c.Join(ch)
                }
            }
        }
    case "433": // nick in use
        // choose a new nick automatically
        oldNick := c.Nick()
        n := oldNick + "_"
        log.Printf("Nick %s is in use, switching to %s", oldNick, n)
        c.setNick(n)
        c.rawf("NICK %s", n)
    case "CAP":
        // server capability negotiation
        // Expect: :server CAP * ACK :sasl
        log.Printf("CAP response: %s %s", strings.Join(args, " "), trailing)
        if len(args) >= 2 && strings.ToUpper(args[1]) == "ACK" && strings.ToLower(args[2]) == "sasl" {
            log.Printf("SASL capability acknowledged, starting authentication")
            c.raw("AUTHENTICATE PLAIN")
        }
    case "AUTHENTICATE":
        // Expect a '+' from server to send payload
        if args[0] == "+" {
            payload := fmt.Sprintf("\x00%s\x00%s", c.saslUser, c.saslPass)
            enc := base64.StdEncoding.EncodeToString([]byte(payload))
            log.Printf("Sending SASL PLAIN credentials")
            c.rawf("AUTHENTICATE %s", enc)
        }
    case "903": // SASL success
        log.Printf("SASL authentication successful")
        c.raw("CAP END")
    case "904", "905": // SASL fail/abort
        log.Printf("SASL authentication failed (code %s)", cmd)
        c.raw("CAP END")
    case "JOIN":
        // :nick!user@host JOIN :#chan
        me := strings.Split(prefix, "!")[0]
        if me == c.Nick() {
            ch := trailing
            if ch == "" && len(args) > 0 {
                ch = args[0]
            }
            if ch != "" {
                log.Printf("Joined channel: %s", ch)
                c.channelsMu.Lock()
                c.channels[strings.ToLower(ch)] = struct{}{}
                c.channelsMu.Unlock()
            }
        }
    case "PART":
        me := strings.Split(prefix, "!")[0]
        if me == c.Nick() && len(args) > 0 {
            ch := args[0]
            log.Printf("Left channel: %s", ch)
            c.channelsMu.Lock()
            delete(c.channels, strings.ToLower(ch))
            c.channelsMu.Unlock()
        }
    case "KICK":
        // :op KICK #chan nick :reason
        if len(args) >= 2 {
            ch, nick := args[0], args[1]
            if nick == c.Nick() {
                log.Printf("Kicked from channel: %s", ch)
                c.channelsMu.Lock()
                delete(c.channels, strings.ToLower(ch))
                c.channelsMu.Unlock()
            }
        }
    case "NICK":
        // :oldnick!u@h NICK :newnick
        me := strings.Split(prefix, "!")[0]
        if me == c.Nick() && trailing != "" {
            log.Printf("Nick changed from %s to %s", c.Nick(), trailing)
            c.setNick(trailing)
        }
    case "PRIVMSG":
        // :sender!user@host PRIVMSG target :message
        log.Printf("PRIVMSG Recv: %s", trailing);
        if len(args) >= 1 && trailing != "" {
            sender := strings.Split(prefix, "!")[0]
            target := args[0]
            message := trailing
            
            // Check if message starts with "@" followed by bot's nick
            botNick := c.Nick()
            mention := "@" + botNick
            if strings.HasPrefix(message, mention) && len(message) > len(mention) && message[len(mention)] == ' ' {
                // Extract the actual message after the mention
                actualMessage := strings.TrimSpace(message[len(mention):])
                log.Printf("Mentioned in %s by %s: %s", target, sender, actualMessage)
                
                // Call n8n webhook if configured
                if c.n8nWebhook != "" {
                    go c.callN8NWebhook(sender, target, actualMessage, message)
                }
            }
        }
    }
}

func (c *IRCClient) callN8NWebhook(sender, target, message, fullMessage string) {
    type N8NPayload struct {
        Sender      string `json:"sender"`
        Target      string `json:"target"`
        Message     string `json:"message"`
        ChatInput string `json:"chatInput"`
        BotNick     string `json:"botNick"`
        SessionId string `json:"sessionId"`
        Timestamp   int64  `json:"timestamp"`
    }

    payload := N8NPayload{
        Sender:      sender,
        Target:      target,
        Message:     message,
        SessionId:  "IRC",
        ChatInput: fullMessage,
        BotNick:     c.Nick(),
        Timestamp:   time.Now().Unix(),
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshaling n8n payload: %v", err)
        return
    }
    log.Printf("Calling webhook: %s", c.n8nWebhook);
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Post(c.n8nWebhook, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error calling n8n webhook: %v", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        log.Printf("Successfully called n8n webhook for message from %s", sender)
    } else {
        log.Printf("n8n webhook returned status %d", resp.StatusCode)
    }
}

func (c *IRCClient) rawf(format string, a ...any) { c.raw(fmt.Sprintf(format, a...)) }

func (c *IRCClient) raw(s string) {
    c.wmu.Lock()
    log.Printf(">> %s", s)
    fmt.Fprint(c.rw, s, "\r\n")
    c.rw.Flush()
    c.wmu.Unlock()
}

func (c *IRCClient) Join(channel string) { c.rawf("JOIN %s", channel) }
func (c *IRCClient) Part(channel string, reason string) {
    if reason == "" {
        c.rawf("PART %s", channel)
    } else {
        c.rawf("PART %s :%s", channel, reason)
    }
}
func (c *IRCClient) Privmsg(target, msg string) { c.rawf("PRIVMSG %s :%s", target, msg) }
func (c *IRCClient) Notice(target, msg string) { c.rawf("NOTICE %s :%s", target, msg) }
func (c *IRCClient) SetNick(n string)           { c.rawf("NICK %s", n) }

func (c *IRCClient) Channels() []string {
    c.channelsMu.RLock()
    defer c.channelsMu.RUnlock()
    out := make([]string, 0, len(c.channels))
    for ch := range c.channels {
        out = append(out, ch)
    }
    return out
}

func (c *IRCClient) Close() error {
    log.Printf("Closing IRC connection")
    if c.conn != nil {
        _ = c.conn.Close()
    }
    c.alive.Store(false)
    return nil
}

// --- Supervisor with reconnect ---

type Supervisor struct {
    client *IRCClient
    stop   chan struct{}
}

func NewSupervisor(c *IRCClient) *Supervisor {
    return &Supervisor{client: c, stop: make(chan struct{})}
}

func (s *Supervisor) Run() {
    backoff := time.Second
    max := 2 * time.Minute

    s.client.onReady = func() {
        log.Printf("Connected as %s", s.client.Nick())
        backoff = time.Second
    }

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
            // Give the connection time to register before checking if it's alive
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
    bot   *IRCClient
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

    mux.HandleFunc("/api/state", a.auth(func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, 200, map[string]any{
            "connected": a.bot.Connected(),
            "nick":      a.bot.Nick(),
            "channels":  a.bot.Channels(),
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

    mux.HandleFunc("/api/notice", a.auth(func(w http.ResponseWriter, r *http.Request) {
        var in struct{ Target, Message string }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Target == "" || in.Message == "" {
            writeJSON(w, 400, errorResponse{"target and message required"})
            return
        }
        a.bot.Notice(in.Target, in.Message)
        writeJSON(w, 200, map[string]string{"status": "ok"})
    }))

    mux.HandleFunc("/api/raw", a.auth(func(w http.ResponseWriter, r *http.Request) {
        var in struct{ Line string }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Line) == "" {
            writeJSON(w, 400, errorResponse{"line required"})
            return
        }
        a.bot.raw(in.Line)
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

func main() {
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)

    apiToken := os.Getenv("API_TOKEN")
    apiAddr := getenv("API_ADDR", ":8080")
    apiTLS := boolenv("API_TLS", false)
    apiCert := os.Getenv("API_CERT")
    apiKey := os.Getenv("API_KEY")

    // Validate TLS configuration
    if apiTLS && (apiCert == "" || apiKey == "") {
        log.Fatalf("API_CERT and API_KEY are required when API_TLS=1")
    }

    bot := NewIRCClient()
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

