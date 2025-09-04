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
    "regexp"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

// Characters that should be ignored when surrounding the bot nick
var ignoreChars = []string{"/"}

// Helper function to check if a nick mention should be ignored based on surrounding characters
func shouldIgnoreNickMention(message, quotedNick string) bool {
    for _, char := range ignoreChars {
        // Check for patterns like /nick/ (both sides)
        bothSidesPattern := char + quotedNick + char
        if matched, _ := regexp.MatchString("(?i)"+regexp.QuoteMeta(bothSidesPattern), message); matched {
            return true
        }
        
        // Check for patterns like /nick (left side only)
        if matched, _ := regexp.MatchString("(?i)"+regexp.QuoteMeta(char)+quotedNick+`\b`, message); matched {
            return true
        }
        
        // Check for patterns like nick/ (right side only)  
        if matched, _ := regexp.MatchString("(?i)"+`\b`+quotedNick+regexp.QuoteMeta(char), message); matched {
            return true
        }
    }
    return false
}

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

type TriggerPayload struct {
    EventType   string            `json:"eventType"`
    Sender      string            `json:"sender"`
    Target      string            `json:"target"`
    Message     string            `json:"message"`
    ChatInput   string            `json:"chatInput"`
    BotNick     string            `json:"botNick"`
    SessionId   string            `json:"sessionId"`
    Timestamp   int64             `json:"timestamp"`
    MessageTags map[string]string `json:"messageTags,omitempty"`
}

// ChannelUser represents a user in a channel with their modes
type ChannelUser struct {
    Nick  string `json:"nick"`
    Modes string `json:"modes,omitempty"` // e.g. "ov" for operator+voice
}

// ChannelState represents the state of an IRC channel
type ChannelState struct {
    Name  string                 `json:"name"`
    Users map[string]string      `json:"users"` // nick -> modes (e.g. "ov", "o", "v", or "" for no modes)
}

// UserModeChange represents a mode change operation
type UserModeChange struct {
    Adding bool   // true for +, false for -
    Mode   rune   // the mode character (o, v, etc.)
    Nick   string // the user affected
}

// PendingRequest represents a request waiting for IRC response
type PendingRequest struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"` // "list" or "whois"
    Target    string    `json:"target,omitempty"` // for whois, the nick being queried
    Data      []map[string]string `json:"data"`
    Complete  bool      `json:"complete"`
    StartTime time.Time `json:"start_time"`
    done      chan bool
}

// WhoisInfo represents collected WHOIS information
type WhoisInfo struct {
    Nick     string `json:"nick"`
    User     string `json:"user,omitempty"`
    Host     string `json:"host,omitempty"`
    RealName string `json:"real_name,omitempty"`
    Server   string `json:"server,omitempty"`
    ServerInfo string `json:"server_info,omitempty"`
    Operator bool   `json:"operator"`
    Idle     string `json:"idle,omitempty"`
    Channels string `json:"channels,omitempty"`
}

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
    triggerConfig TriggerConfig

    conn   net.Conn
    rw     *bufio.ReadWriter
    wmu    sync.Mutex
    alive  atomic.Bool

    channelsMu sync.RWMutex
    channels   map[string]struct{}
    
    // Channel state tracking
    channelStatesMu sync.RWMutex
    channelStates   map[string]*ChannelState // channel name (lowercase) -> state

    // SASL state tracking
    saslInProgress atomic.Bool
    saslComplete   chan bool

    // Pending requests tracking (for LIST and WHOIS)
    pendingMu sync.RWMutex
    pending   map[string]*PendingRequest // request ID -> request

    onReady func()
}

type TriggerConfig struct {
    Endpoints map[string]TriggerEndpoint `json:"endpoints"`
}

type TriggerEndpoint struct {
    URL       string   `json:"url"`
    Token     string   `json:"token"`
    Events    []string `json:"events"`
    Channels  []string `json:"channels,omitempty"`
    Users     []string `json:"users,omitempty"`
}

func NewIRCClient() *IRCClient {
    c := &IRCClient{
        addr:        getenv("IRC_ADDR", ""),
        useTLS:      boolenv("IRC_TLS", true),
        tlsInsecure: boolenv("IRC_TLS_INSECURE", false),
        pass:        os.Getenv("IRC_PASS"),
        user:        getenv("IRC_USER", "Hanna"),
        name:        getenv("IRC_NAME", "Hanna"),
        saslUser:    os.Getenv("SASL_USER"),
        saslPass:    os.Getenv("SASL_PASS"),
        channels:    make(map[string]struct{}),
        channelStates: make(map[string]*ChannelState),
        saslComplete: make(chan bool, 1),
        pending:     make(map[string]*PendingRequest),
    }
    c.nick.Store(getenv("IRC_NICK", "Hanna"))
    
    // Load trigger configuration
    c.loadTriggerConfig()
    
    return c
}

func (c *IRCClient) loadTriggerConfig() {
    configStr := os.Getenv("TRIGGER_CONFIG")
    if configStr == "" {
        // Fallback to legacy N8N_WEBHOOK for backward compatibility
        if webhook := os.Getenv("N8N_WEBHOOK"); webhook != "" {
            c.triggerConfig = TriggerConfig{
                Endpoints: map[string]TriggerEndpoint{
                    "legacy": {
                        URL:    webhook,
                        Events: []string{"mention"},
                    },
                },
            }
        }
        return
    }
    log.Printf("Got trigger config");
    if err := json.Unmarshal([]byte(configStr), &c.triggerConfig); err != nil {
        log.Printf("Error parsing TRIGGER_CONFIG: %v", err)
        c.triggerConfig = TriggerConfig{Endpoints: make(map[string]TriggerEndpoint)}
    }
}

func (c *IRCClient) Connected() bool { return c.alive.Load() }

func (c *IRCClient) Nick() string { return c.nick.Load().(string) }

func (c *IRCClient) setNick(n string) { c.nick.Store(n) }

// Helper functions for channel state tracking
func (c *IRCClient) addUserToChannel(channel, nick string, modes string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    if c.channelStates[channel] == nil {
        c.channelStates[channel] = &ChannelState{
            Name:  channel,
            Users: make(map[string]string),
        }
    }
    c.channelStates[channel].Users[nick] = modes
}

func (c *IRCClient) removeUserFromChannel(channel, nick string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    if state := c.channelStates[channel]; state != nil {
        delete(state.Users, nick)
    }
}

func (c *IRCClient) removeUserFromAllChannels(nick string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    for _, state := range c.channelStates {
        delete(state.Users, nick)
    }
}

func (c *IRCClient) clearChannelState(channel string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    delete(c.channelStates, channel)
}

// parseModeChange parses IRC mode strings like "+oo nick1 nick2" or "-v nick"
func (c *IRCClient) parseModeChange(channel, modeString string, params []string) []UserModeChange {
    var changes []UserModeChange
    adding := true
    paramIdx := 0
    
    for _, char := range modeString {
        switch char {
        case '+':
            adding = true
        case '-':
            adding = false
        case 'o', 'v', 'h', 'b', 'k', 'l': // modes that take parameters
            if paramIdx < len(params) {
                changes = append(changes, UserModeChange{
                    Adding: adding,
                    Mode:   char,
                    Nick:   params[paramIdx],
                })
                paramIdx++
            }
        }
    }
    
    return changes
}

func (c *IRCClient) applyModeChanges(channel string, changes []UserModeChange) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    if state := c.channelStates[channel]; state != nil {
        for _, change := range changes {
            if change.Mode == 'o' || change.Mode == 'v' || change.Mode == 'h' {
                currentModes := state.Users[change.Nick]
                if change.Adding {
                    // Add mode if not present
                    if !strings.ContainsRune(currentModes, change.Mode) {
                        currentModes += string(change.Mode)
                    }
                } else {
                    // Remove mode if present
                    currentModes = strings.ReplaceAll(currentModes, string(change.Mode), "")
                }
                state.Users[change.Nick] = currentModes
            }
        }
    }
}

func (c *IRCClient) getChannelStates() map[string]map[string]interface{} {
    c.channelStatesMu.RLock()
    defer c.channelStatesMu.RUnlock()
    
    result := make(map[string]map[string]interface{})
    for channelName, state := range c.channelStates {
        users := make(map[string]interface{})
        for nick, modes := range state.Users {
            if modes == "" {
                users[nick] = nil
            } else {
                users[nick] = modes
            }
        }
        result[channelName] = users
    }
    return result
}

// Helper functions for pending requests
func (c *IRCClient) createPendingRequest(reqType, target string) *PendingRequest {
    c.pendingMu.Lock()
    defer c.pendingMu.Unlock()
    
    req := &PendingRequest{
        ID:        fmt.Sprintf("%s_%d", reqType, time.Now().UnixNano()),
        Type:      reqType,
        Target:    target,
        Data:      make([]map[string]string, 0),
        Complete:  false,
        StartTime: time.Now(),
        done:      make(chan bool, 1),
    }
    
    c.pending[req.ID] = req
    
    // Cleanup old requests after 30 seconds
    go func() {
        select {
        case <-req.done:
            // Request completed normally
        case <-time.After(30 * time.Second):
            // Request timed out
            c.pendingMu.Lock()
            delete(c.pending, req.ID)
            c.pendingMu.Unlock()
            req.Complete = true
            close(req.done)
        }
    }()
    
    return req
}

func (c *IRCClient) getPendingRequest(id string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    return c.pending[id]
}

func (c *IRCClient) completePendingRequest(id string) {
    c.pendingMu.Lock()
    defer c.pendingMu.Unlock()
    
    if req := c.pending[id]; req != nil {
        req.Complete = true
        select {
        case req.done <- true:
        default:
        }
        delete(c.pending, id)
    }
}

func (c *IRCClient) findPendingRequestByType(reqType string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    
    for _, req := range c.pending {
        if req.Type == reqType && !req.Complete {
            return req
        }
    }
    return nil
}

func (c *IRCClient) findPendingWhoisRequest(nick string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    
    for _, req := range c.pending {
        if req.Type == "whois" && strings.EqualFold(req.Target, nick) && !req.Complete {
            return req
        }
    }
    return nil
}

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

    // Check if SASL is configured
    sasl := c.saslUser != "" && c.saslPass != ""
    
    // Always request CAP negotiation for message-tags (and SASL if configured)
    log.Printf("Starting capability negotiation")
    c.raw("CAP LS 302")
    
    if sasl {
        log.Printf("Requesting SASL and message-tags capabilities")
        c.saslInProgress.Store(true)
        c.raw("CAP REQ :sasl message-tags")
    } else {
        log.Printf("Requesting message-tags capability")
        c.raw("CAP REQ :message-tags")
    }

    go c.readLoop()

    if sasl {
        // Wait for SASL to complete before sending NICK/USER
        log.Printf("Waiting for SASL authentication to complete...")
        select {
        case success := <-c.saslComplete:
            if success {
                log.Printf("SASL authentication completed successfully")
            } else {
                log.Printf("SASL authentication failed, continuing without SASL")
            }
        case <-time.After(30 * time.Second):
            log.Printf("SASL authentication timed out, continuing without SASL")
            c.saslInProgress.Store(false)
        }
    }

    // Send NICK and USER after SASL is complete (or if SASL is not used)
    log.Printf("Sending NICK and USER commands")
    c.rawf("NICK %s", c.Nick())
    c.rawf("USER %s 0 * :%s", c.user, c.name)

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
    // Parse message tags if present
    var tags map[string]string
    rest := line
    
    if strings.HasPrefix(rest, "@") {
        // Find the end of the tags section
        if i := strings.Index(rest, " "); i != -1 {
            tagSection := rest[1:i] // Remove the @ prefix
            rest = strings.TrimSpace(rest[i+1:])
            
            // Parse individual tags
            tags = make(map[string]string)
            if tagSection != "" {
                tagPairs := strings.Split(tagSection, ";")
                for _, pair := range tagPairs {
                    if strings.Contains(pair, "=") {
                        parts := strings.SplitN(pair, "=", 2)
                        key := parts[0]
                        value := parts[1]
                        // Unescape tag values according to IRC v3.2 spec
                        value = strings.ReplaceAll(value, "\\:", ";")
                        value = strings.ReplaceAll(value, "\\s", " ")
                        value = strings.ReplaceAll(value, "\\\\", "\\")
                        value = strings.ReplaceAll(value, "\\r", "\r")
                        value = strings.ReplaceAll(value, "\\n", "\n")
                        tags[key] = value
                    } else {
                        // Tag without value
                        tags[pair] = ""
                    }
                }
            }
            
            if len(tags) > 0 {
                log.Printf("Parsed message tags: %v", tags)
            }
        }
    }
    
    prefix := ""
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
        // Expect: :server CAP * ACK :sasl or :server CAP * ACK sasl
        log.Printf("CAP response: %s %s", strings.Join(args, " "), trailing)
        if len(args) >= 2 && strings.ToUpper(args[1]) == "ACK" {
            capList := trailing
            if capList == "" && len(args) > 2 {
                capList = strings.Join(args[2:], " ")
            }
            log.Printf("Server acknowledged capabilities: %s", capList)
            
            if strings.Contains(strings.ToLower(capList), "message-tags") {
                log.Printf("Message-tags capability enabled")
            }
            
            if strings.Contains(strings.ToLower(capList), "sasl") {
                log.Printf("SASL capability acknowledged, starting authentication")
                c.raw("AUTHENTICATE PLAIN")
            } else if !c.saslInProgress.Load() {
                // No SASL requested, end CAP negotiation
                log.Printf("Ending capability negotiation")
                c.raw("CAP END")
            }
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
        if c.saslInProgress.Load() {
            c.saslInProgress.Store(false)
            select {
            case c.saslComplete <- true:
            default:
            }
        }
    case "904", "905": // SASL fail/abort
        log.Printf("SASL authentication failed (code %s)", cmd)
        c.raw("CAP END")
        if c.saslInProgress.Load() {
            c.saslInProgress.Store(false)
            select {
            case c.saslComplete <- false:
            default:
            }
        }
    case "KICK":
        // :op KICK #chan nick :reason
        if len(args) >= 2 {
            ch, kickedNick := args[0], args[1]
            kicker := strings.Split(prefix, "!")[0]
            reason := trailing
            
            if strings.ToLower(kickedNick) == strings.ToLower(c.Nick()) {
                log.Printf("Kicked from channel: %s", ch)
                c.channelsMu.Lock()
                delete(c.channels, strings.ToLower(ch))
                c.channelsMu.Unlock()
                
                // Clear channel state when we're kicked
                c.clearChannelState(ch)
            } else {
                log.Printf("User %s kicked %s from %s: %s", kicker, kickedNick, ch, reason)
                c.removeUserFromChannel(ch, kickedNick)
                c.sendTriggerEvent("kick", kicker, ch, fmt.Sprintf("%s kicked %s: %s", kicker, kickedNick, reason), reason, tags)
            }
        }
    case "MODE":
        // :nick!user@host MODE target modestring [params...]
        if len(args) >= 2 {
            setter := strings.Split(prefix, "!")[0]
            target := args[0]
            modeString := args[1]
            params := ""
            if len(args) > 2 {
                params = strings.Join(args[2:], " ")
            }
            
            // If target is a channel (starts with # or &), handle channel modes
            if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
                paramList := []string{}
                if len(args) > 2 {
                    paramList = args[2:]
                }
                
                changes := c.parseModeChange(target, modeString, paramList)
                c.applyModeChanges(target, changes)
                
                // Log the mode changes
                for _, change := range changes {
                    op := "+"
                    if !change.Adding {
                        op = "-"
                    }
                    log.Printf("Mode change by %s: %s%c %s in %s", setter, op, change.Mode, change.Nick, target)
                }
            }
            
            message := fmt.Sprintf("Mode %s %s %s", target, modeString, params)
            log.Printf("Mode change by %s: %s", setter, message)
            c.sendTriggerEvent("mode", setter, target, message, message, tags)
        }
    case "TOPIC":
        // :nick!user@host TOPIC #channel :new topic
        if len(args) >= 1 {
            setter := strings.Split(prefix, "!")[0]
            channel := args[0]
            topic := trailing
            
            message := fmt.Sprintf("Topic for %s set by %s: %s", channel, setter, topic)
            log.Printf("Topic change: %s", message)
            c.sendTriggerEvent("topic", setter, channel, message, topic, tags)
        }
    case "NOTICE":
        // :sender!user@host NOTICE target :message
        if len(args) >= 1 && trailing != "" {
            sender := strings.Split(prefix, "!")[0]
            target := args[0]
            message := trailing
            
            log.Printf("NOTICE from %s to %s: %s", sender, target, message)
            c.sendTriggerEvent("notice", sender, target, message, message, tags)
        }
    case "NICK":
        // :oldnick!u@h NICK :newnick
        oldNick := strings.Split(prefix, "!")[0]
        newNick := trailing
        
        if strings.ToLower(oldNick) == strings.ToLower(c.Nick()) && newNick != "" {
            log.Printf("Nick changed from %s to %s", c.Nick(), newNick)
            c.setNick(newNick)
        }
        
        // Update nick in all channel states
        if newNick != "" && oldNick != "" {
            c.channelStatesMu.Lock()
            for _, state := range c.channelStates {
                if modes, exists := state.Users[oldNick]; exists {
                    delete(state.Users, oldNick)
                    state.Users[newNick] = modes
                }
            }
            c.channelStatesMu.Unlock()
        }
    case "PRIVMSG":
        // :sender!user@host PRIVMSG target :message
        log.Printf("PRIVMSG Recv: %s", trailing);
        if len(tags) > 0 {
            log.Printf("Message tags: %v", tags)
        }
        if len(args) >= 1 && trailing != "" {
            sender := strings.Split(prefix, "!")[0]
            target := args[0]
            message := trailing
            
            // Send general privmsg event first
            c.sendTriggerEvent("privmsg", sender, target, message, message, tags)
            // Ignore when surrounded by specific characters like '/'
            botNick := c.Nick()
            
            // Create regex pattern that matches bot nick with word boundaries
            quotedNick := regexp.QuoteMeta(strings.ToLower(botNick))
            pattern := `\b` + quotedNick + `\b`
            regex, err := regexp.Compile("(?i)" + pattern)
            if err != nil {
                log.Printf("Error compiling regex for nick matching: %v", err)
                return
            }
            
            // First check if the nick matches as a word
            if regex.MatchString(message) {
                // Additional check: reject if surrounded by ignore characters
                if shouldIgnoreNickMention(message, quotedNick) {
                    // Skip this match - it's surrounded by ignore characters
                    return
                }
                
                                // This is a valid mention
                log.Printf("Nick mentioned in %s by %s: %s", target, sender, message)
                
                // Send mention event to triggers
                c.sendTriggerEvent("mention", sender, target, message, message, tags)
            }
        }
    case "JOIN":
        // :nick!user@host JOIN :#chan
        senderParts := strings.Split(prefix, "!")
        sender := senderParts[0]
        me := sender
        if strings.ToLower(me) == strings.ToLower(c.Nick()) {
            ch := trailing
            if ch == "" && len(args) > 0 {
                ch = args[0]
            }
            if ch != "" {
                log.Printf("Joined channel: %s", ch)
                c.channelsMu.Lock()
                c.channels[strings.ToLower(ch)] = struct{}{}
                c.channelsMu.Unlock()
                
                // Add ourselves to the channel state
                c.addUserToChannel(ch, c.Nick(), "")
                
                // Request NAMES for this channel to get user list
                c.rawf("NAMES %s", ch)
            }
        } else {
            // Someone else joined
            ch := trailing
            if ch == "" && len(args) > 0 {
                ch = args[0]
            }
            if ch != "" {
                log.Printf("User %s joined %s", sender, ch)
                c.addUserToChannel(ch, sender, "")
                c.sendTriggerEvent("join", sender, ch, "", "", tags)
            }
        }
    case "PART":
        senderParts := strings.Split(prefix, "!")
        sender := senderParts[0]
        me := sender
        if strings.ToLower(me) == strings.ToLower(c.Nick()) && len(args) > 0 {
            ch := args[0]
            log.Printf("Left channel: %s", ch)
            c.channelsMu.Lock()
            delete(c.channels, strings.ToLower(ch))
            c.channelsMu.Unlock()
            
            // Clear channel state when we leave
            c.clearChannelState(ch)
        } else if len(args) > 0 {
            // Someone else parted
            ch := args[0]
            reason := trailing
            log.Printf("User %s left %s: %s", sender, ch, reason)
            c.removeUserFromChannel(ch, sender)
            c.sendTriggerEvent("part", sender, ch, reason, reason, tags)
        }
    case "QUIT":
        // :nick!user@host QUIT :reason
        senderParts := strings.Split(prefix, "!")
        sender := senderParts[0]
        reason := trailing
        log.Printf("User %s quit: %s", sender, reason)
        c.removeUserFromAllChannels(sender)
        c.sendTriggerEvent("quit", sender, "", reason, reason, tags)
    case "353": // RPL_NAMREPLY
        // :server 353 nick = #channel :nick1 @nick2 +nick3
        if len(args) >= 3 && trailing != "" {
            channel := args[2]
            names := strings.Fields(trailing)
            
            log.Printf("NAMES reply for %s: %s", channel, trailing)
            
            for _, name := range names {
                modes := ""
                nick := name
                
                // Parse prefix modes (@, +, %, etc.)
                for len(nick) > 0 {
                    switch nick[0] {
                    case '@':
                        modes += "o"
                        nick = nick[1:]
                    case '+':
                        modes += "v"
                        nick = nick[1:]
                    case '%':
                        modes += "h"
                        nick = nick[1:]
                    default:
                        goto done
                    }
                }
                done:
                
                if nick != "" {
                    c.addUserToChannel(channel, nick, modes)
                }
            }
        }
    case "366": // RPL_ENDOFNAMES
        // :server 366 nick #channel :End of NAMES list
        if len(args) >= 2 {
            channel := args[1]
            log.Printf("End of NAMES list for %s", channel)
        }
    case "322": // RPL_LIST - Channel list entry
        // :server 322 nick #channel users :topic
        if len(args) >= 3 {
            if req := c.findPendingRequestByType("list"); req != nil {
                channel := args[1]
                users := args[2]
                topic := trailing
                
                entry := map[string]string{
                    "channel": channel,
                    "users":   users,
                    "topic":   topic,
                }
                req.Data = append(req.Data, entry)
                log.Printf("LIST entry: %s (%s users) - %s", channel, users, topic)
            }
        }
    case "323": // RPL_LISTEND - End of channel list
        // :server 323 nick :End of LIST
        if req := c.findPendingRequestByType("list"); req != nil {
            log.Printf("End of LIST - found %d channels", len(req.Data))
            c.completePendingRequest(req.ID)
        }
    case "311": // RPL_WHOISUSER
        // :server 311 nick target user host * :real_name
        if len(args) >= 5 && trailing != "" {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                entry := map[string]string{
                    "type":      "user",
                    "nick":      targetNick,
                    "user":      args[2],
                    "host":      args[3],
                    "real_name": trailing,
                }
                req.Data = append(req.Data, entry)
                log.Printf("WHOIS user info for %s: %s@%s (%s)", targetNick, args[2], args[3], trailing)
            }
        }
    case "312": // RPL_WHOISSERVER
        // :server 312 nick target server :server_info
        if len(args) >= 3 && trailing != "" {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                entry := map[string]string{
                    "type":        "server",
                    "nick":        targetNick,
                    "server":      args[2],
                    "server_info": trailing,
                }
                req.Data = append(req.Data, entry)
                log.Printf("WHOIS server info for %s: %s (%s)", targetNick, args[2], trailing)
            }
        }
    case "313": // RPL_WHOISOPERATOR
        // :server 313 nick target :privileges
        if len(args) >= 2 && trailing != "" {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                entry := map[string]string{
                    "type":       "operator",
                    "nick":       targetNick,
                    "privileges": trailing,
                }
                req.Data = append(req.Data, entry)
                log.Printf("WHOIS operator info for %s: %s", targetNick, trailing)
            }
        }
    case "317": // RPL_WHOISIDLE
        // :server 317 nick target seconds :seconds idle
        if len(args) >= 3 && trailing != "" {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                entry := map[string]string{
                    "type":    "idle",
                    "nick":    targetNick,
                    "seconds": args[2],
                    "info":    trailing,
                }
                req.Data = append(req.Data, entry)
                log.Printf("WHOIS idle info for %s: %s seconds (%s)", targetNick, args[2], trailing)
            }
        }
    case "318": // RPL_ENDOFWHOIS
        // :server 318 nick target :info
        if len(args) >= 2 {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                log.Printf("End of WHOIS for %s - collected %d entries", targetNick, len(req.Data))
                c.completePendingRequest(req.ID)
            }
        }
    case "319": // RPL_WHOISCHANNELS
        // :server 319 nick target :*( ( '@' / '+' ) <channel> ' ' )
        if len(args) >= 2 && trailing != "" {
            targetNick := args[1]
            if req := c.findPendingWhoisRequest(targetNick); req != nil {
                entry := map[string]string{
                    "type":     "channels",
                    "nick":     targetNick,
                    "channels": trailing,
                }
                req.Data = append(req.Data, entry)
                log.Printf("WHOIS channels for %s: %s", targetNick, trailing)
            }
        }
    }
}


func (c *IRCClient) sendTriggerEvent(eventType, sender, target, message, fullMessage string, tags map[string]string) {
    payload := TriggerPayload{
        EventType:   eventType,
        Sender:      sender,
        Target:      target,
        Message:     message,
        SessionId:   "IRC",
        ChatInput:   fullMessage,
        BotNick:     c.Nick(),
        Timestamp:   time.Now().Unix(),
        MessageTags: tags,
    }

    for endpointName, endpoint := range c.triggerConfig.Endpoints {
        // Check if this endpoint listens for this event type
        found := false
        for _, event := range endpoint.Events {
            if event == eventType {
                found = true
                break
            }
        }
        if !found {
            continue
        }

        // Check channel filter
        if len(endpoint.Channels) > 0 && target != "" {
            found = false
            for _, ch := range endpoint.Channels {
                if strings.EqualFold(ch, target) {
                    found = true
                    break
                }
            }
            if !found {
                continue
            }
        }

        // Check user filter
        if len(endpoint.Users) > 0 && sender != "" {
            found = false
            for _, user := range endpoint.Users {
                if strings.EqualFold(user, sender) {
                    found = true
                    break
                }
            }
            if !found {
                continue
            }
        }

        // Send to this endpoint
        go c.callTriggerEndpoint(endpointName, endpoint, payload)
    }
}

func (c *IRCClient) callTriggerEndpoint(name string, endpoint TriggerEndpoint, payload TriggerPayload) {
    jsonData, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshaling trigger payload for %s: %v", name, err)
        return
    }

    log.Printf("Calling trigger endpoint %s: %s", name, endpoint.URL)
    
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("POST", endpoint.URL, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Error creating request for %s: %v", name, err)
        return
    }
    
    req.Header.Set("Content-Type", "application/json")
    if endpoint.Token != "" {
        req.Header.Set("Authorization", "Bearer "+endpoint.Token)
    }

    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error calling trigger endpoint %s: %v", name, err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        log.Printf("Successfully called trigger endpoint %s for %s event from %s", name, payload.EventType, payload.Sender)
    } else {
        log.Printf("Trigger endpoint %s returned status %d for %s event", name, resp.StatusCode, payload.EventType)
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
func (c *IRCClient) Privmsg(target, msg string) {
    // IRC protocol: max message length is 512 bytes including command, prefix, etc.
    // Safe to use 450 chars for message body
    const maxMsgLen = 450
    // Split on newlines first
    lines := strings.Split(msg, "\n")
    for _, line := range lines {
        // Split long lines into chunks
        for len(line) > 0 {
            chunk := line
            if len(chunk) > maxMsgLen {
                chunk = chunk[:maxMsgLen]
            }
            c.rawf("PRIVMSG %s :%s", target, chunk)
            line = line[len(chunk):]
        }
    }
}
func (c *IRCClient) Notice(target, msg string) { c.rawf("NOTICE %s :%s", target, msg) }
func (c *IRCClient) SetNick(n string)           { c.rawf("NICK %s", n) }

// List initiates a LIST command and returns a request ID to track the response
func (c *IRCClient) List() string {
    req := c.createPendingRequest("list", "")
    c.raw("LIST")
    return req.ID
}

// Whois initiates a WHOIS command for a specific nick and returns a request ID
func (c *IRCClient) Whois(nick string) string {
    req := c.createPendingRequest("whois", nick)
    c.rawf("WHOIS %s", nick)
    return req.ID
}

// GetRequestResult waits for a request to complete and returns the result
func (c *IRCClient) GetRequestResult(requestID string, timeout time.Duration) (*PendingRequest, error) {
    req := c.getPendingRequest(requestID)
    if req == nil {
        return nil, fmt.Errorf("request not found")
    }
    
    if req.Complete {
        return req, nil
    }
    
    // Wait for completion or timeout
    select {
    case <-req.done:
        return req, nil
    case <-time.After(timeout):
        return req, fmt.Errorf("request timed out")
    }
}

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
            "channels":  a.bot.getChannelStates(),
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

    mux.HandleFunc("/api/list", a.auth(func(w http.ResponseWriter, r *http.Request) {
        if !a.bot.Connected() {
            writeJSON(w, 503, errorResponse{"bot not connected"})
            return
        }
        
        requestID := a.bot.List()
        
        // Wait for the result with a 10 second timeout
        result, err := a.bot.GetRequestResult(requestID, 10*time.Second)
        if err != nil {
            writeJSON(w, 500, errorResponse{fmt.Sprintf("list request failed: %v", err)})
            return
        }
        
        writeJSON(w, 200, map[string]interface{}{
            "channels": result.Data,
            "count":    len(result.Data),
        })
    }))

    mux.HandleFunc("/api/whois", a.auth(func(w http.ResponseWriter, r *http.Request) {
        if !a.bot.Connected() {
            writeJSON(w, 503, errorResponse{"bot not connected"})
            return
        }
        
        var in struct{ Nick string `json:"nick"` }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Nick) == "" {
            writeJSON(w, 400, errorResponse{"nick required"})
            return
        }
        
        requestID := a.bot.Whois(in.Nick)
        
        // Wait for the result with a 10 second timeout
        result, err := a.bot.GetRequestResult(requestID, 10*time.Second)
        if err != nil {
            writeJSON(w, 500, errorResponse{fmt.Sprintf("whois request failed: %v", err)})
            return
        }
        
        // Parse the whois data into a structured format
        whoisInfo := make(map[string]interface{})
        whoisInfo["nick"] = in.Nick
        whoisInfo["raw_data"] = result.Data
        
        // Parse structured data
        for _, entry := range result.Data {
            switch entry["type"] {
            case "user":
                whoisInfo["user"] = entry["user"]
                whoisInfo["host"] = entry["host"]
                whoisInfo["real_name"] = entry["real_name"]
            case "server":
                whoisInfo["server"] = entry["server"]
                whoisInfo["server_info"] = entry["server_info"]
            case "operator":
                whoisInfo["operator"] = true
                whoisInfo["privileges"] = entry["privileges"]
            case "idle":
                whoisInfo["idle_seconds"] = entry["seconds"]
                whoisInfo["idle_info"] = entry["info"]
            case "channels":
                whoisInfo["channels"] = entry["channels"]
            }
        }
        
        writeJSON(w, 200, whoisInfo)
    }))

    a.mux = mux
    return mux
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)

    apiToken := os.Getenv("API_TOKEN")
    apiAddr := getenv("API_ADDR", ":"+getenv("API_PORT", "8080"))
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
