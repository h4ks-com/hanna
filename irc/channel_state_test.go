package irc

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestChannelStateTracking(t *testing.T) {
	// Create a new IRC client for testing
	client := NewClient()
	client.setNick("TestBot")

	t.Run("AddUserToChannel", func(t *testing.T) {
		// Test adding users to a channel
		client.AddUserToChannel("#test", "user1", "")
		client.AddUserToChannel("#test", "user2", "o")
		client.AddUserToChannel("#test", "user3", "v")

		states := client.GetChannelStates()
		if len(states) != 1 {
			t.Errorf("Expected 1 channel, got %d", len(states))
		}

		testChannel, exists := states["#test"]
		if !exists {
			t.Fatal("Channel #test not found in states")
		}

		expected := map[string]interface{}{
			"user1": nil,
			"user2": "o",
			"user3": "v",
		}

		if !reflect.DeepEqual(testChannel, expected) {
			t.Errorf("Expected %v, got %v", expected, testChannel)
		}
	})

	t.Run("RemoveUserFromChannel", func(t *testing.T) {
		// Remove a user from the channel
		client.RemoveUserFromChannel("#test", "user2")

		states := client.GetChannelStates()
		testChannel := states["#test"]

		expected := map[string]interface{}{
			"user1": nil,
			"user3": "v",
		}

		if !reflect.DeepEqual(testChannel, expected) {
			t.Errorf("Expected %v, got %v", expected, testChannel)
		}
	})

	t.Run("RemoveUserFromAllChannels", func(t *testing.T) {
		// Add user to multiple channels
		client.AddUserToChannel("#test2", "user1", "o")
		client.AddUserToChannel("#test2", "user4", "")

		// Remove user1 from all channels
		client.RemoveUserFromAllChannels("user1")

		states := client.GetChannelStates()
		
		// user1 should be gone from both channels
		if _, exists := states["#test"]["user1"]; exists {
			t.Error("user1 should be removed from #test")
		}
		if _, exists := states["#test2"]["user1"]; exists {
			t.Error("user1 should be removed from #test2")
		}

		// Other users should remain
		if _, exists := states["#test"]["user3"]; !exists {
			t.Error("user3 should still be in #test")
		}
		if _, exists := states["#test2"]["user4"]; !exists {
			t.Error("user4 should still be in #test2")
		}
	})

	t.Run("ClearChannelState", func(t *testing.T) {
		// Clear entire channel
		client.ClearChannelState("#test")

		states := client.GetChannelStates()
		if _, exists := states["#test"]; exists {
			t.Error("Channel #test should be cleared")
		}

		// Other channel should remain
		if _, exists := states["#test2"]; !exists {
			t.Error("Channel #test2 should still exist")
		}
	})
}

func TestModeChangeParsing(t *testing.T) {
	client := NewClient()

	testCases := []struct {
		name       string
		modeString string
		params     []string
		expected   []UserModeChange
	}{
		{
			name:       "Grant operator to two users",
			modeString: "+oo",
			params:     []string{"user1", "user2"},
			expected: []UserModeChange{
				{Adding: true, Mode: 'o', Nick: "user1"},
				{Adding: true, Mode: 'o', Nick: "user2"},
			},
		},
		{
			name:       "Remove voice from user",
			modeString: "-v",
			params:     []string{"user3"},
			expected: []UserModeChange{
				{Adding: false, Mode: 'v', Nick: "user3"},
			},
		},
		{
			name:       "Mixed mode changes",
			modeString: "+o-v+h",
			params:     []string{"user1", "user2", "user3"},
			expected: []UserModeChange{
				{Adding: true, Mode: 'o', Nick: "user1"},
				{Adding: false, Mode: 'v', Nick: "user2"},
				{Adding: true, Mode: 'h', Nick: "user3"},
			},
		},
		{
			name:       "Mode with no parameters",
			modeString: "+t",
			params:     []string{},
			expected:   nil, // No user mode changes for channel modes
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			changes := client.ParseModeChange("#test", tc.modeString, tc.params)
			
			// Handle nil vs empty slice comparison
			if tc.expected == nil && len(changes) == 0 {
				return // Both are effectively empty
			}
			
			if !reflect.DeepEqual(changes, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, changes)
			}
		})
	}
}

func TestApplyModeChanges(t *testing.T) {
	client := NewClient()

	// Set up initial channel state
	client.AddUserToChannel("#test", "user1", "")
	client.AddUserToChannel("#test", "user2", "v")
	client.AddUserToChannel("#test", "user3", "o")

	t.Run("Grant operator mode", func(t *testing.T) {
		changes := []UserModeChange{
			{Adding: true, Mode: 'o', Nick: "user1"},
		}
		client.ApplyModeChanges("#test", changes)

		states := client.GetChannelStates()
		if states["#test"]["user1"] != "o" {
			t.Errorf("Expected user1 to have 'o' mode, got %v", states["#test"]["user1"])
		}
	})

	t.Run("Grant voice to operator", func(t *testing.T) {
		changes := []UserModeChange{
			{Adding: true, Mode: 'v', Nick: "user3"},
		}
		client.ApplyModeChanges("#test", changes)

		states := client.GetChannelStates()
		modes := states["#test"]["user3"].(string)
		if !containsMode(modes, 'o') || !containsMode(modes, 'v') {
			t.Errorf("Expected user3 to have both 'o' and 'v' modes, got %v", modes)
		}
	})

	t.Run("Remove voice mode", func(t *testing.T) {
		changes := []UserModeChange{
			{Adding: false, Mode: 'v', Nick: "user2"},
		}
		client.ApplyModeChanges("#test", changes)

		states := client.GetChannelStates()
		if states["#test"]["user2"] != nil && states["#test"]["user2"] != "" {
			t.Errorf("Expected user2 to have no modes, got %v", states["#test"]["user2"])
		}
	})

	t.Run("Remove operator from user with multiple modes", func(t *testing.T) {
		changes := []UserModeChange{
			{Adding: false, Mode: 'o', Nick: "user3"},
		}
		client.ApplyModeChanges("#test", changes)

		states := client.GetChannelStates()
		modes := states["#test"]["user3"].(string)
		if containsMode(modes, 'o') {
			t.Errorf("Expected user3 to not have 'o' mode, got %v", modes)
		}
		if !containsMode(modes, 'v') {
			t.Errorf("Expected user3 to still have 'v' mode, got %v", modes)
		}
	})
}

func TestIRCCommandHandling(t *testing.T) {
	client := NewClient()
	client.setNick("TestBot")

	t.Run("Handle JOIN command manually", func(t *testing.T) {
		// Manually simulate the effects of a JOIN command without calling handleLine
		channel := "#test"
		nick := "TestBot"
		
		// Add to channels map
		client.channelsMu.Lock()
		client.channels[strings.ToLower(channel)] = struct{}{}
		client.channelsMu.Unlock()
		
		// Add to channel state
		client.AddUserToChannel(channel, nick, "")

		// Check if we're in the channel list
		client.channelsMu.RLock()
		_, exists := client.channels["#test"]
		client.channelsMu.RUnlock()

		if !exists {
			t.Error("Bot should be in #test channel")
		}

		// Check if we're in the channel state
		states := client.GetChannelStates()
		if _, exists := states["#test"]["TestBot"]; !exists {
			t.Error("TestBot should be in #test channel state")
		}
	})

	t.Run("Handle other user JOIN manually", func(t *testing.T) {
		client.AddUserToChannel("#test", "user1", "")

		states := client.GetChannelStates()
		if _, exists := states["#test"]["user1"]; !exists {
			t.Error("user1 should be in #test channel state")
		}
	})

	t.Run("Handle NAMES reply parsing", func(t *testing.T) {
		// Clear existing state
		client.ClearChannelState("#test")
		
		// Manually simulate NAMES processing
		channel := "#test"
		names := []string{"TestBot", "@user1", "+user2", "user3"}
		
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
				client.AddUserToChannel(channel, nick, modes)
			}
		}

		states := client.GetChannelStates()
		testChannel := states["#test"]

		expected := map[string]interface{}{
			"TestBot": nil,
			"user1":   "o",
			"user2":   "v",
			"user3":   nil,
		}

		if !reflect.DeepEqual(testChannel, expected) {
			t.Errorf("Expected %v, got %v", expected, testChannel)
		}
	})

	t.Run("Handle MODE command manually", func(t *testing.T) {
		// Manually apply mode change
		modeString := "+o"
		params := []string{"user3"}
		
		changes := client.ParseModeChange("#test", modeString, params)
		client.ApplyModeChanges("#test", changes)

		states := client.GetChannelStates()
		if states["#test"]["user3"] != "o" {
			t.Errorf("Expected user3 to have 'o' mode, got %v", states["#test"]["user3"])
		}
	})

	t.Run("Handle PART command manually", func(t *testing.T) {
		client.RemoveUserFromChannel("#test", "user2")

		states := client.GetChannelStates()
		if _, exists := states["#test"]["user2"]; exists {
			t.Error("user2 should not be in #test after PART")
		}
	})

	t.Run("Handle QUIT command manually", func(t *testing.T) {
		// Add user1 to another channel first
		client.AddUserToChannel("#test2", "user1", "v")
		
		// User quits - remove from all channels
		client.RemoveUserFromAllChannels("user1")

		states := client.GetChannelStates()
		
		// user1 should be removed from all channels
		if _, exists := states["#test"]["user1"]; exists {
			t.Error("user1 should not be in #test after QUIT")
		}
		if _, exists := states["#test2"]["user1"]; exists {
			t.Error("user1 should not be in #test2 after QUIT")
		}
	})

	t.Run("Handle NICK change manually", func(t *testing.T) {
		// Manually simulate nick change
		oldNick := "user3"
		newNick := "newuser3"
		
		client.channelStatesMu.Lock()
		for _, state := range client.channelStates {
			if modes, exists := state.Users[oldNick]; exists {
				delete(state.Users, oldNick)
				state.Users[newNick] = modes
			}
		}
		client.channelStatesMu.Unlock()

		states := client.GetChannelStates()
		
		// Old nick should be gone
		if _, exists := states["#test"]["user3"]; exists {
			t.Error("user3 should not exist after nick change")
		}
		
		// New nick should exist with same modes
		if states["#test"]["newuser3"] != "o" {
			t.Errorf("Expected newuser3 to have 'o' mode, got %v", states["#test"]["newuser3"])
		}
	})
}

func TestAPIStateEndpoint(t *testing.T) {
	client := NewClient()
	client.setNick("TestBot")

	// Set up some channel state
	client.AddUserToChannel("#lobby", "Valware", "")
	client.AddUserToChannel("#lobby", "handyc", "o")
	client.AddUserToChannel("#lobby", "mattf", "o")
	client.AddUserToChannel("#bots", "Hanna", "")
	client.AddUserToChannel("#bots", "Samantha", "vo")

	states := client.GetChannelStates()

	// Verify the format matches the expected JSON structure
	expectedLobby := map[string]interface{}{
		"Valware": nil,
		"handyc":  "o",
		"mattf":   "o",
	}

	expectedBots := map[string]interface{}{
		"Hanna":    nil,
		"Samantha": "vo",
	}

	if !reflect.DeepEqual(states["#lobby"], expectedLobby) {
		t.Errorf("Expected lobby %v, got %v", expectedLobby, states["#lobby"])
	}

	if !reflect.DeepEqual(states["#bots"], expectedBots) {
		t.Errorf("Expected bots %v, got %v", expectedBots, states["#bots"])
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(map[string]interface{}{
		"connected": true,
		"nick":      "TestBot",
		"channels":  states,
	})
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify structure is correct
	if result["connected"] != true {
		t.Error("Expected connected to be true")
	}
	if result["nick"] != "TestBot" {
		t.Error("Expected nick to be TestBot")
	}

	channels, ok := result["channels"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected channels to be a map")
	}

	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}
}

// Helper function to check if a mode string contains a specific mode
func containsMode(modes string, mode rune) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}
