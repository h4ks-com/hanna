package main

import (
	"testing"
)

// Test IRC message handling for LIST and WHOIS responses
func TestListMessageHandling(t *testing.T) {
	client := NewIRCClient()
	
	// Create a pending LIST request
	req := client.createPendingRequest("list", "")
	
	// Simulate IRC LIST responses
	// 322 response: channel list entry
	client.handleLine(":irc.server.com 322 botname #general 42 :General discussion channel")
	client.handleLine(":irc.server.com 322 botname #bots 15 :Bot testing area")
	client.handleLine(":irc.server.com 322 botname #support 8 :Technical support")
	
	// 323 response: end of list
	client.handleLine(":irc.server.com 323 botname :End of LIST")
	
	// Verify the request is complete
	if !req.Complete {
		t.Error("LIST request should be complete after 323 response")
	}
	
	// Verify the collected data
	if len(req.Data) != 3 {
		t.Errorf("Expected 3 channel entries, got %d", len(req.Data))
	}
	
	// Check first channel entry
	if len(req.Data) > 0 {
		entry := req.Data[0]
		if entry["channel"] != "#general" {
			t.Errorf("Expected channel '#general', got '%s'", entry["channel"])
		}
		if entry["users"] != "42" {
			t.Errorf("Expected users '42', got '%s'", entry["users"])
		}
		if entry["topic"] != "General discussion channel" {
			t.Errorf("Expected topic 'General discussion channel', got '%s'", entry["topic"])
		}
	}
}

func TestWhoisMessageHandling(t *testing.T) {
	client := NewIRCClient()
	
	// Create a pending WHOIS request
	req := client.createPendingRequest("whois", "testuser")
	
	// Simulate IRC WHOIS responses
	// 311: RPL_WHOISUSER
	client.handleLine(":irc.server.com 311 botname testuser johndoe example.com * :John Doe")
	
	// 312: RPL_WHOISSERVER  
	client.handleLine(":irc.server.com 312 botname testuser irc.libera.chat :Stockholm, Sweden")
	
	// 313: RPL_WHOISOPERATOR
	client.handleLine(":irc.server.com 313 botname testuser :is an IRC operator")
	
	// 317: RPL_WHOISIDLE
	client.handleLine(":irc.server.com 317 botname testuser 42 :seconds idle")
	
	// 319: RPL_WHOISCHANNELS
	client.handleLine(":irc.server.com 319 botname testuser :@#ops +#general #random")
	
	// 318: RPL_ENDOFWHOIS
	client.handleLine(":irc.server.com 318 botname testuser :End of WHOIS list")
	
	// Verify the request is complete
	if !req.Complete {
		t.Error("WHOIS request should be complete after 318 response")
	}
	
	// Verify the collected data
	if len(req.Data) != 5 {
		t.Errorf("Expected 5 WHOIS entries, got %d", len(req.Data))
	}
	
	// Check entries by type
	foundTypes := make(map[string]bool)
	for _, entry := range req.Data {
		entryType := entry["type"]
		foundTypes[entryType] = true
		
		switch entryType {
		case "user":
			if entry["nick"] != "testuser" {
				t.Errorf("Expected nick 'testuser', got '%s'", entry["nick"])
			}
			if entry["user"] != "johndoe" {
				t.Errorf("Expected user 'johndoe', got '%s'", entry["user"])
			}
			if entry["host"] != "example.com" {
				t.Errorf("Expected host 'example.com', got '%s'", entry["host"])
			}
			if entry["real_name"] != "John Doe" {
				t.Errorf("Expected real_name 'John Doe', got '%s'", entry["real_name"])
			}
		case "server":
			if entry["server"] != "irc.libera.chat" {
				t.Errorf("Expected server 'irc.libera.chat', got '%s'", entry["server"])
			}
			if entry["server_info"] != "Stockholm, Sweden" {
				t.Errorf("Expected server_info 'Stockholm, Sweden', got '%s'", entry["server_info"])
			}
		case "operator":
			if entry["privileges"] != "is an IRC operator" {
				t.Errorf("Expected privileges 'is an IRC operator', got '%s'", entry["privileges"])
			}
		case "idle":
			if entry["seconds"] != "42" {
				t.Errorf("Expected seconds '42', got '%s'", entry["seconds"])
			}
			if entry["info"] != "seconds idle" {
				t.Errorf("Expected info 'seconds idle', got '%s'", entry["info"])
			}
		case "channels":
			if entry["channels"] != "@#ops +#general #random" {
				t.Errorf("Expected channels '@#ops +#general #random', got '%s'", entry["channels"])
			}
		}
	}
	
	// Verify all expected types were found
	expectedTypes := []string{"user", "server", "operator", "idle", "channels"}
	for _, expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Missing WHOIS entry type: %s", expectedType)
		}
	}
}

func TestWhoisWithoutOptionalFields(t *testing.T) {
	client := NewIRCClient()
	
	// Create a pending WHOIS request
	req := client.createPendingRequest("whois", "regularuser")
	
	// Simulate minimal WHOIS response (user without operator status, idle info, etc.)
	client.handleLine(":irc.server.com 311 botname regularuser jane example.org * :Jane Smith")
	client.handleLine(":irc.server.com 312 botname regularuser irc.libera.chat :Stockholm, Sweden")
	client.handleLine(":irc.server.com 319 botname regularuser :#general #help")
	client.handleLine(":irc.server.com 318 botname regularuser :End of WHOIS list")
	
	// Verify the request is complete
	if !req.Complete {
		t.Error("WHOIS request should be complete after 318 response")
	}
	
	// Should have 3 entries (user, server, channels)
	if len(req.Data) != 3 {
		t.Errorf("Expected 3 WHOIS entries, got %d", len(req.Data))
	}
}

func TestCaseInsensitiveWhoisMatching(t *testing.T) {
	client := NewIRCClient()
	
	// Create a pending WHOIS request with lowercase nick
	req := client.createPendingRequest("whois", "testuser")
	
	// Simulate IRC response with different case
	client.handleLine(":irc.server.com 311 botname TestUser johndoe example.com * :John Doe")
	client.handleLine(":irc.server.com 318 botname TestUser :End of WHOIS list")
	
	// Should still match and complete the request
	if !req.Complete {
		t.Error("WHOIS request should complete with case-insensitive nick matching")
	}
	
	if len(req.Data) != 1 {
		t.Errorf("Expected 1 WHOIS entry, got %d", len(req.Data))
	}
}

func TestListWithoutPendingRequest(t *testing.T) {
	client := NewIRCClient()
	
	// Try to handle LIST responses without a pending request
	// This should not cause any errors or panics
	client.handleLine(":irc.server.com 322 botname #channel 10 :Some channel")
	client.handleLine(":irc.server.com 323 botname :End of LIST")
	
	// Should handle gracefully without pending request
	// No assertions needed, just ensuring no panic
}

func TestWhoisWithoutPendingRequest(t *testing.T) {
	client := NewIRCClient()
	
	// Try to handle WHOIS responses without a pending request
	// This should not cause any errors or panics
	client.handleLine(":irc.server.com 311 botname someuser user host * :Real Name")
	client.handleLine(":irc.server.com 318 botname someuser :End of WHOIS list")
	
	// Should handle gracefully without pending request
	// No assertions needed, just ensuring no panic
}
