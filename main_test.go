package main

import (
	"log"
	"os"
	"strings"
	"testing"
)

// TestNickMatching tests the actual PRIVMSG nick matching functionality
func TestNickMatching(t *testing.T) {
	testNicks := []string{"Hanna", "hanna-test", "my_hanna"}
	
	for _, botNick := range testNicks {
		t.Run("nick_"+botNick, func(t *testing.T) {
			// Create a test IRC client
			client := &IRCClient{
				channels: make(map[string]struct{}),
			}
			client.nick.Store(botNick)
			
			// Capture log output to detect when nick is mentioned
			// We'll check if the "Nick mentioned" log is printed
			var logOutput strings.Builder
			log.SetOutput(&logOutput)
			defer log.SetOutput(os.Stderr)
			
			testCases := []struct {
				message  string
				expected bool
				desc     string
			}{
				// Should match (should log "Nick mentioned")
				{botNick + " hello", true, "nick at start"},
				{"hello " + botNick, true, "nick at end"},
				{"hey " + botNick + " how are you", true, "nick in middle"},
				{"@" + botNick + " test", true, "nick with @ prefix"},
				{strings.ToUpper(botNick) + " hello", true, "case insensitive match"},
				{strings.ToLower(botNick) + " test", true, "lowercase match"},
				{"Hi, " + botNick + "!", true, "nick with punctuation"},
				
				// Should NOT match (should NOT log "Nick mentioned")
				{botNick + "ah says hi", false, "partial match in word"},
				{botNick + "Bot is here", false, "nick as part of longer word"},
				{"The " + botNick + "Love channel", false, "nick as prefix of word"},
				{"/" + botNick + "/", false, "nick surrounded by slashes"},
				{"test/" + botNick + "/bot", false, "nick with slashes around"},
				{"/" + botNick, false, "nick with slash on left"},
				{botNick + "/", false, "nick with slash on right"},
				{"some" + botNick, false, "nick as suffix"},
				{botNick + "lore", false, "nick as prefix of different word"},
				{"nothing here", false, "no nick mention"},
			}
			
			for _, tc := range testCases {
				t.Run(tc.desc, func(t *testing.T) {
					// Clear log buffer
					logOutput.Reset()
					
					// Create a PRIVMSG line that handleLine would receive
					// Format: :sender!user@host PRIVMSG #channel :message
					line := ":testuser!user@host PRIVMSG #testchan :" + tc.message
					
					// Call the actual handleLine function
					client.handleLine(line)
					
					// Check if "Nick mentioned" appears in log output
					logStr := logOutput.String()
					nickMentioned := strings.Contains(logStr, "Nick mentioned")
					
					if nickMentioned != tc.expected {
						t.Errorf("Message: %q - Expected nick mentioned: %v, Got: %v (log: %s)", tc.message, tc.expected, nickMentioned, logStr)
					}
				})
			}
		})
	}
}

