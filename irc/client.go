// Package irc provides IRC client functionality for Hanna bot
package irc

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
    "regexp"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "time"
)

const Version = "2.0.0"

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
    Name         string            `json:"name"`
    Users        map[string]string `json:"users"`         // nick -> modes (e.g. "ov", "o", "v", or "" for no modes)
    Topic        string            `json:"topic"`         // current topic
    TopicSetBy   string            `json:"topic_set_by"`  // who set the topic
    TopicSetTime int64             `json:"topic_set_time"` // when topic was set (unix timestamp)
    CreatedTime  int64             `json:"created_time"`  // channel creation time (unix timestamp)
    Modes        string            `json:"modes"`         // channel modes (e.g. "+nt")
    ModeParams   []string          `json:"mode_params"`   // parameters for modes that take them
    BanList      []BanListEntry    `json:"ban_list"`      // channel ban list
    InviteList   []InviteListEntry `json:"invite_list"`   // channel invite list
    ExceptList   []ExceptListEntry `json:"except_list"`   // channel exception list
    URL          string            `json:"url,omitempty"` // channel URL if set
    SpecialInfo  map[string]string `json:"special_info,omitempty"` // other special channel info
}

// BanListEntry represents a ban list entry
type BanListEntry struct {
    Mask      string `json:"mask"`
    SetBy     string `json:"set_by"`
    SetTime   int64  `json:"set_time"`
    Reason    string `json:"reason,omitempty"`
}

// InviteListEntry represents an invite list entry
type InviteListEntry struct {
    Mask      string `json:"mask"`
    SetBy     string `json:"set_by"`
    SetTime   int64  `json:"set_time"`
}

// ExceptListEntry represents an exception list entry
type ExceptListEntry struct {
    Mask      string `json:"mask"`
    SetBy     string `json:"set_by"`
    SetTime   int64  `json:"set_time"`
}

// ServerInfo represents information about the IRC server
type ServerInfo struct {
    Name         string            `json:"name"`
    Version      string            `json:"version"`
    UserModes    string            `json:"user_modes"`
    ChannelModes string            `json:"channel_modes"`
    Created      string            `json:"created"`
    ISupportTags map[string]string `json:"isupport_tags"` // RPL_ISUPPORT (005) tags
    AdminInfo    AdminInfo         `json:"admin_info"`
    MOTD         []string          `json:"motd"`
    LocalUsers   int               `json:"local_users"`
    GlobalUsers  int               `json:"global_users"`
    MaxLocalUsers int              `json:"max_local_users"`
    MaxGlobalUsers int             `json:"max_global_users"`
    Operators    int               `json:"operators"`
    UnknownConns int               `json:"unknown_connections"`
    Channels     int               `json:"channels_formed"`
}

// AdminInfo represents server administrator information
type AdminInfo struct {
    Server   string `json:"server"`
    Location1 string `json:"location1"`
    Location2 string `json:"location2"`
    Email    string `json:"email"`
}

// UserInfo represents comprehensive information about a user
type UserInfo struct {
    Nick           string            `json:"nick"`
    User           string            `json:"user"`
    Host           string            `json:"host"`
    RealName       string            `json:"real_name"`
    Server         string            `json:"server"`
    ServerInfo     string            `json:"server_info"`
    Account        string            `json:"account,omitempty"`        // services account (330)
    IsOperator     bool              `json:"is_operator"`             // global IRC operator
    IsAway         bool              `json:"is_away"`
    AwayMessage    string            `json:"away_message,omitempty"`
    IdleTime       int               `json:"idle_time"`               // seconds idle
    SignonTime     int64             `json:"signon_time"`             // when they signed on
    Channels       []string          `json:"channels"`                // list of channels
    Modes          string            `json:"modes"`                   // user modes
    IsSecure       bool              `json:"is_secure"`               // using SSL/TLS (671)
    CertFingerprint string           `json:"cert_fingerprint,omitempty"` // SSL cert fingerprint (276)
    ActualHost     string            `json:"actual_host,omitempty"`   // actual host/IP (338)
    CountryCode    string            `json:"country_code,omitempty"`  // country code (344)
    Country        string            `json:"country,omitempty"`       // country name (344)
    ASN            string            `json:"asn,omitempty"`           // AS number (569)
    IsBot          bool              `json:"is_bot"`                  // marked as bot (335)
    WebIRCGateway  string            `json:"webirc_gateway,omitempty"` // WebIRC gateway info
    SpecialInfo    map[string]string `json:"special_info,omitempty"`  // other special info
}

// StatEntry represents a server statistics entry
type StatEntry struct {
    Type   string            `json:"type"`
    Data   map[string]string `json:"data"`
}

// IRCError represents an IRC error response
type IRCError struct {
    Code    string `json:"code"`
    Target  string `json:"target,omitempty"`
    Message string `json:"message"`
    Time    int64  `json:"time"`
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

type Client struct {
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

    // User information tracking
    userInfoMu sync.RWMutex
    userInfo   map[string]*UserInfo // nick (lowercase) -> user info

    // Server information tracking
    serverInfoMu sync.RWMutex
    serverInfo   *ServerInfo

    // Statistics tracking
    statsMu sync.RWMutex
    stats   []StatEntry

    // Error tracking (recent errors)
    errorsMu sync.RWMutex
    errors   []IRCError

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

func NewClient() *Client {
    c := &Client{
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
        userInfo:     make(map[string]*UserInfo),
        serverInfo:   &ServerInfo{ISupportTags: make(map[string]string)},
        stats:        make([]StatEntry, 0),
        errors:       make([]IRCError, 0),
        saslComplete: make(chan bool, 1),
        pending:     make(map[string]*PendingRequest),
    }
    c.nick.Store(getenv("IRC_NICK", "Hanna"))
    
    // Load trigger configuration
    c.loadTriggerConfig()
    
    return c
}

func (c *Client) loadTriggerConfig() {
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
        log.Fatalf("FATAL: Invalid TRIGGER_CONFIG JSON: %v", err)
    }
}

func (c *Client) Connected() bool { return c.alive.Load() }

func (c *Client) Nick() string { return c.nick.Load().(string) }

func (c *Client) setNick(n string) { c.nick.Store(n) }

// Helper functions for channel state tracking
func (c *Client) AddUserToChannel(channel, nick string, modes string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    if c.channelStates[channel] == nil {
        c.channelStates[channel] = &ChannelState{
            Name:        channel,
            Users:       make(map[string]string),
            BanList:     make([]BanListEntry, 0),
            InviteList:  make([]InviteListEntry, 0),
            ExceptList:  make([]ExceptListEntry, 0),
            ModeParams:  make([]string, 0),
            SpecialInfo: make(map[string]string),
        }
    }
    c.channelStates[channel].Users[nick] = modes
}

func (c *Client) RemoveUserFromChannel(channel, nick string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    if state := c.channelStates[channel]; state != nil {
        delete(state.Users, nick)
    }
}

func (c *Client) RemoveUserFromAllChannels(nick string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    for _, state := range c.channelStates {
        delete(state.Users, nick)
    }
}

func (c *Client) ClearChannelState(channel string) {
    c.channelStatesMu.Lock()
    defer c.channelStatesMu.Unlock()
    
    channel = strings.ToLower(channel)
    delete(c.channelStates, channel)
}

// Helper functions for user information tracking
func (c *Client) updateUserInfo(nick string, updateFunc func(*UserInfo)) {
    c.userInfoMu.Lock()
    defer c.userInfoMu.Unlock()
    
    nick = strings.ToLower(nick)
    if c.userInfo[nick] == nil {
        c.userInfo[nick] = &UserInfo{
            Nick:        nick,
            SpecialInfo: make(map[string]string),
        }
    }
    updateFunc(c.userInfo[nick])
}

func (c *Client) getUserInfo(nick string) *UserInfo {
    c.userInfoMu.RLock()
    defer c.userInfoMu.RUnlock()
    
    nick = strings.ToLower(nick)
    if info := c.userInfo[nick]; info != nil {
        // Return a copy to avoid race conditions
        copyInfo := *info
        copySpecial := make(map[string]string)
        for k, v := range info.SpecialInfo {
            copySpecial[k] = v
        }
        copyInfo.SpecialInfo = copySpecial
        return &copyInfo
    }
    return nil
}

func (c *Client) removeUserInfo(nick string) {
    c.userInfoMu.Lock()
    defer c.userInfoMu.Unlock()
    
    nick = strings.ToLower(nick)
    delete(c.userInfo, nick)
}

// Helper functions for server information tracking
func (c *Client) updateServerInfo(updateFunc func(*ServerInfo)) {
    c.serverInfoMu.Lock()
    defer c.serverInfoMu.Unlock()
    updateFunc(c.serverInfo)
}

func (c *Client) getServerInfo() *ServerInfo {
    c.serverInfoMu.RLock()
    defer c.serverInfoMu.RUnlock()
    
    // Return a copy
    info := *c.serverInfo
    
    // Deep copy maps
    info.ISupportTags = make(map[string]string)
    for k, v := range c.serverInfo.ISupportTags {
        info.ISupportTags[k] = v
    }
    
    // Copy MOTD slice
    info.MOTD = make([]string, len(c.serverInfo.MOTD))
    copy(info.MOTD, c.serverInfo.MOTD)
    
    return &info
}

// Helper functions for statistics tracking
func (c *Client) addStatEntry(statType string, data map[string]string) {
    c.statsMu.Lock()
    defer c.statsMu.Unlock()
    
    c.stats = append(c.stats, StatEntry{
        Type: statType,
        Data: data,
    })
    
    // Keep only the last 1000 stat entries to prevent memory growth
    if len(c.stats) > 1000 {
        c.stats = c.stats[len(c.stats)-1000:]
    }
}

func (c *Client) getStats() []StatEntry {
    c.statsMu.RLock()
    defer c.statsMu.RUnlock()
    
    // Return a copy
    stats := make([]StatEntry, len(c.stats))
    copy(stats, c.stats)
    return stats
}

// Helper functions for error tracking
func (c *Client) addError(code, target, message string) {
    c.errorsMu.Lock()
    defer c.errorsMu.Unlock()
    
    c.errors = append(c.errors, IRCError{
        Code:    code,
        Target:  target,
        Message: message,
        Time:    time.Now().Unix(),
    })
    
    // Keep only the last 100 errors to prevent memory growth
    if len(c.errors) > 100 {
        c.errors = c.errors[len(c.errors)-100:]
    }
}

func (c *Client) getRecentErrors() []IRCError {
    c.errorsMu.RLock()
    defer c.errorsMu.RUnlock()
    
    // Return a copy
    errors := make([]IRCError, len(c.errors))
    copy(errors, c.errors)
    return errors
}

func (c *Client) getAllUsers() map[string]*UserInfo {
    c.userInfoMu.RLock()
    defer c.userInfoMu.RUnlock()
    
    users := make(map[string]*UserInfo)
    for nick, info := range c.userInfo {
        // Create a copy
        copyInfo := *info
        copySpecial := make(map[string]string)
        for k, v := range info.SpecialInfo {
            copySpecial[k] = v
        }
        copyInfo.SpecialInfo = copySpecial
        users[nick] = &copyInfo
    }
    return users
}

// parseTime parses various time formats used in IRC
func parseIRCTime(timeStr string) int64 {
    // Try parsing as Unix timestamp first
    if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
        return timestamp
    }
    
    // Try parsing common IRC date formats
    formats := []string{
        time.RFC1123,
        time.RFC1123Z,
        "Mon Jan 2 15:04:05 2006",
        "Mon Jan 2 15:04:05 MST 2006",
        "Jan 2 15:04:05 2006",
        "2006-01-02 15:04:05",
    }
    
    for _, format := range formats {
        if t, err := time.Parse(format, timeStr); err == nil {
            return t.Unix()
        }
    }
    
    return time.Now().Unix() // fallback to current time
}

// extractServerName extracts server name from IRC prefix
func extractServerName(prefix string) string {
    if prefix == "" {
        return ""
    }
    
    // If prefix contains '!' it's a user, extract server differently
    if strings.Contains(prefix, "!") {
        return ""
    }
    
    // If it's just a server name, return it
    return prefix
}

// parseModeChange parses IRC mode strings like "+oo nick1 nick2" or "-v nick"
func (c *Client) ParseModeChange(channel, modeString string, params []string) []UserModeChange {
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

func (c *Client) ApplyModeChanges(channel string, changes []UserModeChange) {
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

func (c *Client) GetChannelStates() map[string]map[string]interface{} {
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
func (c *Client) createPendingRequest(reqType, target string) *PendingRequest {
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

func (c *Client) getPendingRequest(id string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    return c.pending[id]
}

func (c *Client) completePendingRequest(id string) {
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

func (c *Client) findPendingRequestByType(reqType string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    
    for _, req := range c.pending {
        if req.Type == reqType && !req.Complete {
            return req
        }
    }
    return nil
}

func (c *Client) findPendingWhoisRequest(nick string) *PendingRequest {
    c.pendingMu.RLock()
    defer c.pendingMu.RUnlock()
    
    for _, req := range c.pending {
        if req.Type == "whois" && strings.EqualFold(req.Target, nick) && !req.Complete {
            return req
        }
    }
    return nil
}

func (c *Client) Dial(ctx context.Context) error {
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

func (c *Client) readLoop() {
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

func (c *Client) handleLine(line string) {
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
        // set bot mode +B-)
        c.rawf("MODE %s +B", c.Nick())
        log.Printf("Setting bot mode (+B)")
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
        c.addError(cmd, oldNick, trailing) // Add error tracking
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
        // Track this in user info for our own nick
        c.updateUserInfo(c.Nick(), func(info *UserInfo) {
            if info.SpecialInfo == nil {
                info.SpecialInfo = make(map[string]string)
            }
            info.SpecialInfo["sasl_authenticated"] = "true"
        })
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
        c.addError(cmd, "", trailing) // Add error tracking
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
                c.ClearChannelState(ch)
            } else {
                log.Printf("User %s kicked %s from %s: %s", kicker, kickedNick, ch, reason)
                c.RemoveUserFromChannel(ch, kickedNick)
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
                
                changes := c.ParseModeChange(target, modeString, paramList)
                c.ApplyModeChanges(target, changes)
                
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
                c.AddUserToChannel(ch, c.Nick(), "")
                
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
                c.AddUserToChannel(ch, sender, "")
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
            c.ClearChannelState(ch)
        } else if len(args) > 0 {
            // Someone else parted
            ch := args[0]
            reason := trailing
            log.Printf("User %s left %s: %s", sender, ch, reason)
            c.RemoveUserFromChannel(ch, sender)
            c.sendTriggerEvent("part", sender, ch, reason, reason, tags)
        }
    case "QUIT":
        // :nick!user@host QUIT :reason
        senderParts := strings.Split(prefix, "!")
        sender := senderParts[0]
        reason := trailing
        log.Printf("User %s quit: %s", sender, reason)
        c.RemoveUserFromAllChannels(sender)
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
                    c.AddUserToChannel(channel, nick, modes)
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
    // RFC1459 and Extended IRC Numerics - Comprehensive State Tracking
    case "002": // RPL_YOURHOST
        // :server 002 nick :Your host is servername, running version
        c.updateServerInfo(func(info *ServerInfo) {
            info.Name = extractServerName(prefix)
            // Parse version from message if possible
            if versionMatch := regexp.MustCompile(`running version (.+)$`).FindStringSubmatch(trailing); len(versionMatch) > 1 {
                info.Version = versionMatch[1]
            }
        })
    case "003": // RPL_CREATED
        // :server 003 nick :This server was created date
        c.updateServerInfo(func(info *ServerInfo) {
            if createdMatch := regexp.MustCompile(`created (.+)$`).FindStringSubmatch(trailing); len(createdMatch) > 1 {
                info.Created = createdMatch[1]
            }
        })
    case "004": // RPL_MYINFO
        // :server 004 nick servername version usermodes chanmodes [chanmodes_with_param]
        if len(args) >= 4 {
            c.updateServerInfo(func(info *ServerInfo) {
                info.Name = args[1]
                info.Version = args[2]
                info.UserModes = args[3]
                if len(args) > 4 {
                    info.ChannelModes = args[4]
                }
            })
        }
    case "005": // RPL_ISUPPORT
        // :server 005 nick TOKEN1=value TOKEN2 :are supported by this server
        if len(args) >= 2 {
            c.updateServerInfo(func(info *ServerInfo) {
                // Parse all args except the first (nick) and last (usually a description)
                for i := 1; i < len(args); i++ {
                    token := args[i]
                    if strings.Contains(token, "=") {
                        parts := strings.SplitN(token, "=", 2)
                        info.ISupportTags[parts[0]] = parts[1]
                    } else {
                        info.ISupportTags[token] = ""
                    }
                }
            })
        }
    case "251": // RPL_LUSERCLIENT
        // :server 251 nick :There are <int> users and <int> invisible on <int> servers
        if matches := regexp.MustCompile(`(\d+) users and (\d+) invisible on (\d+) servers`).FindStringSubmatch(trailing); len(matches) > 3 {
            if totalUsers, err := strconv.Atoi(matches[1]); err == nil {
                if invisible, err := strconv.Atoi(matches[2]); err == nil {
                    c.updateServerInfo(func(info *ServerInfo) {
                        info.GlobalUsers = totalUsers + invisible
                    })
                }
            }
        }
    case "252": // RPL_LUSEROP
        // :server 252 nick <int> :operator(s) online
        if len(args) >= 2 {
            if ops, err := strconv.Atoi(args[1]); err == nil {
                c.updateServerInfo(func(info *ServerInfo) {
                    info.Operators = ops
                })
            }
        }
    case "253": // RPL_LUSERUNKNOWN
        // :server 253 nick <int> :unknown connection(s)
        if len(args) >= 2 {
            if unknown, err := strconv.Atoi(args[1]); err == nil {
                c.updateServerInfo(func(info *ServerInfo) {
                    info.UnknownConns = unknown
                })
            }
        }
    case "254": // RPL_LUSERCHANNELS
        // :server 254 nick <int> :channels formed
        if len(args) >= 2 {
            if channels, err := strconv.Atoi(args[1]); err == nil {
                c.updateServerInfo(func(info *ServerInfo) {
                    info.Channels = channels
                })
            }
        }
    case "255": // RPL_LUSERME
        // :server 255 nick :I have <int> clients and <int> servers
        if matches := regexp.MustCompile(`I have (\d+) clients and (\d+) servers`).FindStringSubmatch(trailing); len(matches) > 2 {
            if clients, err := strconv.Atoi(matches[1]); err == nil {
                c.updateServerInfo(func(info *ServerInfo) {
                    info.LocalUsers = clients
                })
            }
        }
    case "256": // RPL_ADMINME
        // :server 256 nick server :Administrative info
        if len(args) >= 2 {
            c.updateServerInfo(func(info *ServerInfo) {
                info.AdminInfo.Server = args[1]
            })
        }
    case "257": // RPL_ADMINLOC1
        // :server 257 nick :admin_location
        c.updateServerInfo(func(info *ServerInfo) {
            info.AdminInfo.Location1 = trailing
        })
    case "258": // RPL_ADMINLOC2
        // :server 258 nick :admin_location
        c.updateServerInfo(func(info *ServerInfo) {
            info.AdminInfo.Location2 = trailing
        })
    case "259": // RPL_ADMINEMAIL
        // :server 259 nick :email_address
        c.updateServerInfo(func(info *ServerInfo) {
            info.AdminInfo.Email = trailing
        })
    case "265": // RPL_LOCALUSERS
        // :server 265 nick [<u> <m>] :Current local users <u>, max <m>
        if len(args) >= 3 {
            if current, err := strconv.Atoi(args[1]); err == nil {
                if max, err := strconv.Atoi(args[2]); err == nil {
                    c.updateServerInfo(func(info *ServerInfo) {
                        info.LocalUsers = current
                        info.MaxLocalUsers = max
                    })
                }
            }
        }
    case "266": // RPL_GLOBALUSERS
        // :server 266 nick [<u> <m>] :Current global users <u>, max <m>
        if len(args) >= 3 {
            if current, err := strconv.Atoi(args[1]); err == nil {
                if max, err := strconv.Atoi(args[2]); err == nil {
                    c.updateServerInfo(func(info *ServerInfo) {
                        info.GlobalUsers = current
                        info.MaxGlobalUsers = max
                    })
                }
            }
        }
    case "276": // RPL_WHOISCERTFP
        // :server 276 nick target :has client certificate fingerprint fingerprint
        if len(args) >= 2 {
            targetNick := args[1]
            if fpMatch := regexp.MustCompile(`fingerprint (.+)$`).FindStringSubmatch(trailing); len(fpMatch) > 1 {
                c.updateUserInfo(targetNick, func(info *UserInfo) {
                    info.CertFingerprint = fpMatch[1]
                })
            }
        }
    case "301": // RPL_AWAY
        // :server 301 nick target :away_message
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.IsAway = true
                info.AwayMessage = trailing
            })
        }
    case "305": // RPL_UNAWAY
        // :server 305 nick :info
        c.updateUserInfo(c.Nick(), func(info *UserInfo) {
            info.IsAway = false
            info.AwayMessage = ""
        })
    case "306": // RPL_NOWAWAY
        // :server 306 nick :info
        c.updateUserInfo(c.Nick(), func(info *UserInfo) {
            info.IsAway = true
        })
    case "307": // RPL_WHOISREGNICK / RPL_WHOISSERVICE
        // :server 307 nick target :info
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["registered"] = "true"
            })
        }
    case "308": // RPL_WHOISADMIN / RPL_RULESSTART
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["admin"] = "true"
            })
        }
    case "309": // RPL_ENDOFRULES / RPL_WHOISSERVICE
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["service"] = "true"
            })
        }
    case "310": // RPL_WHOISHELPOP / RPL_WHOISSERVICE
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["helpop"] = "true"
            })
        }
    case "314": // RPL_WHOWASUSER
        // :server 314 nick target user host * :real_name
        if len(args) >= 5 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.User = args[2]
                info.Host = args[3]
                info.RealName = trailing
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["whowas"] = "true"
            })
        }
    case "320": // RPL_WHOISSPECIAL / RPL_WHOIS_HIDDEN
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["special"] = trailing
            })
        }
    case "324": // RPL_CHANNELMODEIS
        // :server 324 nick channel mode mode_params
        if len(args) >= 3 {
            channel := strings.ToLower(args[1])
            modes := args[2]
            var params []string
            if len(args) > 3 {
                params = args[3:]
            }
            
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].Modes = modes
            c.channelStates[channel].ModeParams = params
            c.channelStatesMu.Unlock()
        }
    case "325": // RPL_UNIQOPIS / RPL_CHANNELPASSIS / RPL_WHOISWEBIRC
        if len(args) >= 3 && strings.HasPrefix(args[1], "#") {
            // Channel related
            channel := strings.ToLower(args[1])
            c.channelStatesMu.Lock()
            if c.channelStates[channel] != nil {
                if c.channelStates[channel].SpecialInfo == nil {
                    c.channelStates[channel].SpecialInfo = make(map[string]string)
                }
                c.channelStates[channel].SpecialInfo["unique_op"] = args[2]
            }
            c.channelStatesMu.Unlock()
        } else if len(args) >= 2 {
            // User related (WHOISWEBIRC)
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["webirc"] = "true"
            })
        }
    case "328": // RPL_CHANNEL_URL
        // :server 328 nick channel :url
        if len(args) >= 2 {
            channel := strings.ToLower(args[1])
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].URL = trailing
            c.channelStatesMu.Unlock()
        }
    case "329": // RPL_CREATIONTIME
        // :server 329 nick channel timestamp
        if len(args) >= 3 {
            channel := strings.ToLower(args[1])
            if timestamp, err := strconv.ParseInt(args[2], 10, 64); err == nil {
                c.channelStatesMu.Lock()
                if c.channelStates[channel] == nil {
                    c.channelStates[channel] = &ChannelState{
                        Name:        channel,
                        Users:       make(map[string]string),
                        BanList:     make([]BanListEntry, 0),
                        InviteList:  make([]InviteListEntry, 0),
                        ExceptList:  make([]ExceptListEntry, 0),
                        ModeParams:  make([]string, 0),
                        SpecialInfo: make(map[string]string),
                    }
                }
                c.channelStates[channel].CreatedTime = timestamp
                c.channelStatesMu.Unlock()
            }
        }
    case "330": // RPL_WHOISACCOUNT / RPL_WHOISLOGGEDIN
        // :server 330 nick target authname :info
        if len(args) >= 3 {
            targetNick := args[1]
            account := args[2]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.Account = account
            })
        }
    case "331": // RPL_NOTOPIC
        // :server 331 nick channel :info
        if len(args) >= 2 {
            channel := strings.ToLower(args[1])
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].Topic = ""
            c.channelStatesMu.Unlock()
        }
    case "332": // RPL_TOPIC
        // :server 332 nick channel :topic
        if len(args) >= 2 {
            channel := strings.ToLower(args[1])
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].Topic = trailing
            c.channelStatesMu.Unlock()
        }
    case "333": // RPL_TOPICWHOTIME
        // :server 333 nick channel nick!user@host timestamp
        if len(args) >= 4 {
            channel := strings.ToLower(args[1])
            topicSetter := args[2]
            if timestamp, err := strconv.ParseInt(args[3], 10, 64); err == nil {
                c.channelStatesMu.Lock()
                if c.channelStates[channel] == nil {
                    c.channelStates[channel] = &ChannelState{
                        Name:        channel,
                        Users:       make(map[string]string),
                        BanList:     make([]BanListEntry, 0),
                        InviteList:  make([]InviteListEntry, 0),
                        ExceptList:  make([]ExceptListEntry, 0),
                        ModeParams:  make([]string, 0),
                        SpecialInfo: make(map[string]string),
                    }
                }
                c.channelStates[channel].TopicSetBy = topicSetter
                c.channelStates[channel].TopicSetTime = timestamp
                c.channelStatesMu.Unlock()
            }
        }
    case "335": // RPL_WHOISBOT
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.IsBot = true
            })
        }
    case "338": // RPL_WHOISACTUALLY
        // :server 338 nick target :is actually user@host ip
        if len(args) >= 2 {
            targetNick := args[1]
            if actualMatch := regexp.MustCompile(`is actually (.+)`).FindStringSubmatch(trailing); len(actualMatch) > 1 {
                c.updateUserInfo(targetNick, func(info *UserInfo) {
                    info.ActualHost = actualMatch[1]
                })
            }
        }
    case "344": // RPL_WHOISCOUNTRY
        // :server 344 nick target countrycode :is connecting from country
        if len(args) >= 3 {
            targetNick := args[1]
            countryCode := args[2]
            if countryMatch := regexp.MustCompile(`from (.+)$`).FindStringSubmatch(trailing); len(countryMatch) > 1 {
                c.updateUserInfo(targetNick, func(info *UserInfo) {
                    info.CountryCode = countryCode
                    info.Country = countryMatch[1]
                })
            }
        }
    case "346": // RPL_INVITELIST
        // :server 346 nick channel invitemask [who set-ts]
        if len(args) >= 3 {
            channel := strings.ToLower(args[1])
            mask := args[2]
            entry := InviteListEntry{Mask: mask}
            
            if len(args) > 3 {
                entry.SetBy = args[3]
            }
            if len(args) > 4 {
                entry.SetTime = parseIRCTime(args[4])
            }
            
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].InviteList = append(c.channelStates[channel].InviteList, entry)
            c.channelStatesMu.Unlock()
        }
    case "347": // RPL_ENDOFINVITELIST
        if len(args) >= 2 {
            log.Printf("End of invite list for %s", args[1])
        }
    case "348": // RPL_EXCEPTLIST
        // :server 348 nick channel exceptionmask [who set-ts]
        if len(args) >= 3 {
            channel := strings.ToLower(args[1])
            mask := args[2]
            entry := ExceptListEntry{Mask: mask}
            
            if len(args) > 3 {
                entry.SetBy = args[3]
            }
            if len(args) > 4 {
                entry.SetTime = parseIRCTime(args[4])
            }
            
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].ExceptList = append(c.channelStates[channel].ExceptList, entry)
            c.channelStatesMu.Unlock()
        }
    case "349": // RPL_ENDOFEXCEPTLIST
        if len(args) >= 2 {
            log.Printf("End of exception list for %s", args[1])
        }
    case "350": // RPL_WHOISGATEWAY
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.WebIRCGateway = trailing
            })
        }
    case "367": // RPL_BANLIST
        // :server 367 nick channel banid [setter time_left|time_left :reason]
        if len(args) >= 3 {
            channel := strings.ToLower(args[1])
            mask := args[2]
            entry := BanListEntry{Mask: mask}
            
            if len(args) > 3 {
                entry.SetBy = args[3]
            }
            if len(args) > 4 {
                entry.SetTime = parseIRCTime(args[4])
            }
            if trailing != "" && trailing != args[2] {
                entry.Reason = trailing
            }
            
            c.channelStatesMu.Lock()
            if c.channelStates[channel] == nil {
                c.channelStates[channel] = &ChannelState{
                    Name:        channel,
                    Users:       make(map[string]string),
                    BanList:     make([]BanListEntry, 0),
                    InviteList:  make([]InviteListEntry, 0),
                    ExceptList:  make([]ExceptListEntry, 0),
                    ModeParams:  make([]string, 0),
                    SpecialInfo: make(map[string]string),
                }
            }
            c.channelStates[channel].BanList = append(c.channelStates[channel].BanList, entry)
            c.channelStatesMu.Unlock()
        }
    case "368": // RPL_ENDOFBANLIST
        if len(args) >= 2 {
            log.Printf("End of ban list for %s", args[1])
        }
    case "371": // RPL_INFO
        // :server 371 nick :string
        c.updateServerInfo(func(info *ServerInfo) {
            info.MOTD = append(info.MOTD, trailing)
        })
    case "372": // RPL_MOTD
        // :server 372 nick :- string
        c.updateServerInfo(func(info *ServerInfo) {
            // Remove leading "- " from MOTD lines
            line := trailing
            if strings.HasPrefix(line, "- ") {
                line = line[2:]
            }
            info.MOTD = append(info.MOTD, line)
        })
    case "375": // RPL_MOTDSTART
        // :server 375 nick :- server Message of the day -
        c.updateServerInfo(func(info *ServerInfo) {
            info.MOTD = []string{} // Clear existing MOTD
        })
    case "376": // RPL_ENDOFMOTD
        log.Printf("End of MOTD")
    case "378": // RPL_WHOISHOST
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["host_info"] = trailing
            })
        }
    case "379": // RPL_WHOISMODES
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["user_modes"] = trailing
            })
        }
    case "569": // RPL_WHOISASN
        // :server 569 nick target asn :is connecting from AS#### [Org Name]
        if len(args) >= 3 {
            targetNick := args[1]
            asn := args[2]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.ASN = asn
            })
        }
    case "671": // RPL_WHOISSECURE / RPL_WHOISSSL
        // :server 671 nick target :is using a secure connection
        if len(args) >= 2 {
            targetNick := args[1]
            c.updateUserInfo(targetNick, func(info *UserInfo) {
                info.IsSecure = true
            })
        }
    // Error numerics - track for debugging/monitoring
    case "400", "401", "402", "403", "404", "405", "406", "407", "408", "409",
         "410", "411", "412", "413", "414", "415", "416", "417", "421", "422",
         "423", "424", "431", "432", "436", "437", "441", "442", "443",
         "444", "445", "446", "451", "461", "462", "463", "464", "465", "466",
         "467", "471", "472", "473", "474", "475", "476", "477", "478", "481",
         "482", "483", "484", "485", "491", "492", "501", "502":
        target := ""
        if len(args) >= 2 {
            target = args[1]
        }
        c.addError(cmd, target, trailing)
        log.Printf("IRC Error %s: %s", cmd, trailing)
    // SASL Authentication numerics
    case "900": // RPL_LOGGEDIN
        // :server 900 nick nick!ident@host account :You are now logged in as user
        if len(args) >= 3 {
            account := args[2]
            c.updateUserInfo(c.Nick(), func(info *UserInfo) {
                info.Account = account
            })
            log.Printf("SASL: Logged in as %s", account)
        }
    case "901": // RPL_LOGGEDOUT
        // :server 901 nick nick!ident@host :You are now logged out
        c.updateUserInfo(c.Nick(), func(info *UserInfo) {
            info.Account = ""
        })
        log.Printf("SASL: Logged out")
    case "902": // ERR_NICKLOCKED
        c.addError(cmd, c.Nick(), trailing)
        log.Printf("SASL: Nick locked - %s", trailing)
    case "906": // ERR_SASLABORTED
        log.Printf("SASL: Authentication aborted")
    case "907": // ERR_SASLALREADY
        log.Printf("SASL: Already authenticated")
    case "908": // RPL_SASLMECHS
        log.Printf("SASL: Available mechanisms - %s", trailing)
    // Statistics numerics - track for monitoring
    case "211", "212", "213", "214", "215", "216", "217", "218", "219",
         "241", "242", "243", "244", "245", "246", "247", "248", "249", "250":
        statData := make(map[string]string)
        statData["numeric"] = cmd
        statData["raw"] = strings.Join(args, " ") + " :" + trailing
        for i, arg := range args {
            statData[fmt.Sprintf("arg%d", i)] = arg
        }
        if trailing != "" {
            statData["trailing"] = trailing
        }
        c.addStatEntry("stats_"+cmd, statData)
    // Additional user information numerics
    case "396": // RPL_VISIBLEHOST / RPL_YOURDISPLAYEDHOST / RPL_HOSTHIDDEN
        // :server 396 nick hostname :is now your visible host
        if len(args) >= 2 {
            hostname := args[1]
            c.updateUserInfo(c.Nick(), func(info *UserInfo) {
                if info.SpecialInfo == nil {
                    info.SpecialInfo = make(map[string]string)
                }
                info.SpecialInfo["visible_host"] = hostname
            })
        }
    // Default case for unhandled numerics - log for potential future implementation
    default:
        if len(cmd) == 3 && cmd[0] >= '0' && cmd[0] <= '9' {
            log.Printf("Unhandled numeric %s: %s %s", cmd, strings.Join(args, " "), trailing)
            // Store unknown numerics for analysis
            statData := make(map[string]string)
            statData["numeric"] = cmd
            statData["args"] = strings.Join(args, " ")
            statData["trailing"] = trailing
            c.addStatEntry("unknown_numeric", statData)
        }
    }
}


func (c *Client) sendTriggerEvent(eventType, sender, target, message, fullMessage string, tags map[string]string) {
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

func (c *Client) callTriggerEndpoint(name string, endpoint TriggerEndpoint, payload TriggerPayload) {
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

func (c *Client) rawf(format string, a ...any) { c.raw(fmt.Sprintf(format, a...)) }

func (c *Client) raw(s string) {
    c.wmu.Lock()
    log.Printf(">> %s", s)
    fmt.Fprint(c.rw, s, "\r\n")
    c.rw.Flush()
    c.wmu.Unlock()
}

func (c *Client) Join(channel string) { c.rawf("JOIN %s", channel) }
func (c *Client) Part(channel string, reason string) {
    if reason == "" {
        c.rawf("PART %s", channel)
    } else {
        c.rawf("PART %s :%s", channel, reason)
    }
}
func (c *Client) Privmsg(target, msg string) {
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
func (c *Client) Notice(target, msg string) { c.rawf("NOTICE %s :%s", target, msg) }
func (c *Client) SetNick(n string)           { c.rawf("NICK %s", n) }

// List initiates a LIST command and returns a request ID to track the response
func (c *Client) List() string {
    req := c.createPendingRequest("list", "")
    c.raw("LIST")
    return req.ID
}

// Whois initiates a WHOIS command for a specific nick and returns a request ID
func (c *Client) Whois(nick string) string {
    req := c.createPendingRequest("whois", nick)
    c.rawf("WHOIS %s", nick)
    return req.ID
}

// GetRequestResult waits for a request to complete and returns the result
func (c *Client) GetRequestResult(requestID string, timeout time.Duration) (*PendingRequest, error) {
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

func (c *Client) Channels() []string {
    c.channelsMu.RLock()
    defer c.channelsMu.RUnlock()
    out := make([]string, 0, len(c.channels))
    for ch := range c.channels {
        out = append(out, ch)
    }
    return out
}

func (c *Client) Close() error {
    log.Printf("Closing IRC connection")
    if c.conn != nil {
        _ = c.conn.Close()
    }
    c.alive.Store(false)
    return nil
}

// --- Supervisor with reconnect ---

type Supervisor struct {
    client *Client
    stop   chan struct{}
}

func NewSupervisor(c *Client) *Supervisor {
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
    bot   *Client
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

    mux.HandleFunc("/api/server", a.auth(func(w http.ResponseWriter, r *http.Request) {
        serverInfo := a.bot.getServerInfo()
        writeJSON(w, 200, serverInfo)
    }))

    mux.HandleFunc("/api/users", a.auth(func(w http.ResponseWriter, r *http.Request) {
        // Get all user information
        a.bot.userInfoMu.RLock()
        users := make(map[string]*UserInfo)
        for nick, info := range a.bot.userInfo {
            // Create a copy
            copyInfo := *info
            copySpecial := make(map[string]string)
            for k, v := range info.SpecialInfo {
                copySpecial[k] = v
            }
            copyInfo.SpecialInfo = copySpecial
            users[nick] = &copyInfo
        }
        a.bot.userInfoMu.RUnlock()
        
        writeJSON(w, 200, map[string]any{
            "users": users,
            "count": len(users),
        })
    }))

    mux.HandleFunc("/api/user", a.auth(func(w http.ResponseWriter, r *http.Request) {
        var in struct{ Nick string `json:"nick"` }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Nick == "" {
            writeJSON(w, 400, errorResponse{"nick required"})
            return
        }
        
        userInfo := a.bot.getUserInfo(in.Nick)
        if userInfo == nil {
            writeJSON(w, 404, errorResponse{"user not found"})
            return
        }
        
        writeJSON(w, 200, userInfo)
    }))

    mux.HandleFunc("/api/stats", a.auth(func(w http.ResponseWriter, r *http.Request) {
        stats := a.bot.getStats()
        writeJSON(w, 200, map[string]any{
            "stats": stats,
            "count": len(stats),
        })
    }))

    mux.HandleFunc("/api/errors", a.auth(func(w http.ResponseWriter, r *http.Request) {
        errors := a.bot.getRecentErrors()
        writeJSON(w, 200, map[string]any{
            "errors": errors,
            "count":  len(errors),
        })
    }))

    mux.HandleFunc("/api/channel", a.auth(func(w http.ResponseWriter, r *http.Request) {
        var in struct{ Channel string `json:"channel"` }
        if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Channel == "" {
            writeJSON(w, 400, errorResponse{"channel required"})
            return
        }
        
        a.bot.channelStatesMu.RLock()
        channelState := a.bot.channelStates[strings.ToLower(in.Channel)]
        a.bot.channelStatesMu.RUnlock()
        
        if channelState == nil {
            writeJSON(w, 404, errorResponse{"channel not found"})
            return
        }
        
        // Create a copy to avoid race conditions
        stateCopy := *channelState
        stateCopy.Users = make(map[string]string)
        for k, v := range channelState.Users {
            stateCopy.Users[k] = v
        }
        stateCopy.BanList = make([]BanListEntry, len(channelState.BanList))
        copy(stateCopy.BanList, channelState.BanList)
        stateCopy.InviteList = make([]InviteListEntry, len(channelState.InviteList))
        copy(stateCopy.InviteList, channelState.InviteList)
        stateCopy.ExceptList = make([]ExceptListEntry, len(channelState.ExceptList))
        copy(stateCopy.ExceptList, channelState.ExceptList)
        stateCopy.ModeParams = make([]string, len(channelState.ModeParams))
        copy(stateCopy.ModeParams, channelState.ModeParams)
        if channelState.SpecialInfo != nil {
            stateCopy.SpecialInfo = make(map[string]string)
            for k, v := range channelState.SpecialInfo {
                stateCopy.SpecialInfo[k] = v
            }
        }
        
        writeJSON(w, 200, &stateCopy)
    }))

    mux.HandleFunc("/api/comprehensive-state", a.auth(func(w http.ResponseWriter, r *http.Request) {
        // Return comprehensive IRC state information
        writeJSON(w, 200, map[string]any{
            "connected":    a.bot.Connected(),
            "nick":         a.bot.Nick(),
            "server":       a.bot.getServerInfo(),
            "channels":     a.bot.GetChannelStates(),
            "users":        a.bot.getAllUsers(),
            "stats":        a.bot.getStats(),
            "recent_errors": a.bot.getRecentErrors(),
            "timestamp":    time.Now().Unix(),
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

