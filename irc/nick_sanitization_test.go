package irc

import (
	"os"
	"strings"
	"testing"
)

func TestSanitizeNick(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "Hanna",
		},
		{
			name:     "valid simple nick",
			input:    "TestBot",
			expected: "TestBot",
		},
		{
			name:     "valid nick with numbers",
			input:    "Bot123",
			expected: "Bot123",
		},
		{
			name:     "valid nick with allowed special chars",
			input:    "Bot_{test}-[v2]",
			expected: "Bot_{test}-[v2]",
		},
		{
			name:     "valid nick with backtick",
			input:    "Bot`test",
			expected: "Bot`test",
		},
		{
			name:     "valid nick with curly braces",
			input:    "Bot{test}",
			expected: "Bot{test}",
		},
		{
			name:     "nick with invalid characters",
			input:    "Bot@#$%^&*()",
			expected: "Bot",
		},
		{
			name:     "nick with spaces",
			input:    "My Bot Name",
			expected: "MyBotName",
		},
		{
			name:     "nick with unicode characters",
			input:    "BötNämé",
			expected: "BtNm",
		},
		{
			name:     "nick with only invalid characters",
			input:    "@#$%^&*()",
			expected: "Hanna",
		},
		{
			name:     "nick exactly 63 characters",
			input:    strings.Repeat("a", 63),
			expected: strings.Repeat("a", 63),
		},
		{
			name:     "nick over 63 characters",
			input:    strings.Repeat("a", 100),
			expected: strings.Repeat("a", 63),
		},
		{
			name:     "nick with mixed valid and invalid chars over 63",
			input:    strings.Repeat("a", 50) + "@#$%^&*" + strings.Repeat("b", 20),
			expected: strings.Repeat("a", 50) + strings.Repeat("b", 13), // truncated to 63
		},
		{
			name:     "nick with dots and other punctuation",
			input:    "Bot.test@example.com",
			expected: "Bottestexamplecom",
		},
		{
			name:     "nick with all valid special characters",
			input:    "Test_-[]{}` ",
			expected: "Test_-[]{}`",
		},
		{
			name:     "nick starting with number",
			input:    "123Bot",
			expected: "123Bot",
		},
		{
			name:     "nick with only underscores and dashes",
			input:    "___---___",
			expected: "___---___",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeNick(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeNick(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewClientWithSanitizedNick(t *testing.T) {
	// Save original env
	oldNick := os.Getenv("IRC_NICK")
	defer func() {
		os.Setenv("IRC_NICK", oldNick)
	}()
	
	tests := []struct {
		name        string
		envNick     string
		expectedNick string
	}{
		{
			name:        "valid nick from env",
			envNick:     "ValidBot",
			expectedNick: "ValidBot",
		},
		{
			name:        "invalid chars in env nick",
			envNick:     "Bot@#$Name",
			expectedNick: "BotName",
		},
		{
			name:        "empty env nick uses default",
			envNick:     "",
			expectedNick: "Hanna",
		},
		{
			name:        "only invalid chars uses default",
			envNick:     "@#$%^&*()",
			expectedNick: "Hanna",
		},
		{
			name:        "nick too long gets truncated",
			envNick:     strings.Repeat("x", 100),
			expectedNick: strings.Repeat("x", 63),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("IRC_NICK", tt.envNick)
			
			client := NewClient()
			actualNick := client.Nick()
			
			if actualNick != tt.expectedNick {
				t.Errorf("NewClient with IRC_NICK=%q got nick %q, want %q", 
					tt.envNick, actualNick, tt.expectedNick)
			}
		})
	}
}

func TestSetNickSanitization(t *testing.T) {
	client := NewClient()
	
	// Mock the raw function to capture output
	var sentCommands []string
	client.testRawCapture = func(s string) {
		sentCommands = append(sentCommands, s)
	}
	
	tests := []struct {
		name        string
		inputNick   string
		expectedCmd string
	}{
		{
			name:        "valid nick",
			inputNick:   "TestBot",
			expectedCmd: "NICK TestBot",
		},
		{
			name:        "nick with invalid chars",
			inputNick:   "Test@Bot#Name",
			expectedCmd: "NICK TestBotName",
		},
		{
			name:        "empty nick uses default",
			inputNick:   "",
			expectedCmd: "NICK Hanna",
		},
		{
			name:        "only invalid chars uses default",
			inputNick:   "@#$%",
			expectedCmd: "NICK Hanna",
		},
		{
			name:        "nick too long gets truncated",
			inputNick:   strings.Repeat("a", 100),
			expectedCmd: "NICK " + strings.Repeat("a", 63),
		},
		{
			name:        "nick with valid special chars",
			inputNick:   "Bot_test-[v2]{}` ",
			expectedCmd: "NICK Bot_test-[v2]{}`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentCommands = nil // Reset
			
			client.SetNick(tt.inputNick)
			
			if len(sentCommands) != 1 {
				t.Fatalf("Expected 1 command, got %d", len(sentCommands))
			}
			
			if sentCommands[0] != tt.expectedCmd {
				t.Errorf("SetNick(%q) sent %q, want %q", 
					tt.inputNick, sentCommands[0], tt.expectedCmd)
			}
		})
	}
}

func TestValidIRCNickCharacters(t *testing.T) {
	// Test all valid IRC nick characters (keeping under 63 char limit)
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123{}[]_-`"
	
	result := sanitizeNick(validChars)
	if result != validChars {
		t.Errorf("Valid IRC characters should not be modified. Got %q (len=%d), want %q (len=%d)", 
			result, len(result), validChars, len(validChars))
	}
}

func TestInvalidIRCNickCharacters(t *testing.T) {
	// Test various invalid characters
	invalidChars := "!@#$%^&*()+=|\\:;\"'<>?,./ \t\n\r"
	
	result := sanitizeNick("test" + invalidChars + "nick")
	expected := "testnick"
	
	if result != expected {
		t.Errorf("Invalid IRC characters should be removed. Got %q, want %q", result, expected)
	}
}

func TestNickLengthLimits(t *testing.T) {
	// Test exactly at the boundary
	exactLimit := strings.Repeat("a", 63)
	result := sanitizeNick(exactLimit)
	if result != exactLimit {
		t.Errorf("Nick of exactly 63 chars should not be truncated. Got len=%d, want len=%d", 
			len(result), len(exactLimit))
	}
	
	// Test over the boundary
	overLimit := strings.Repeat("b", 64)
	result = sanitizeNick(overLimit)
	expected := strings.Repeat("b", 63)
	if result != expected {
		t.Errorf("Nick over 63 chars should be truncated. Got len=%d, want len=%d", 
			len(result), len(expected))
	}
	if len(result) != 63 {
		t.Errorf("Truncated nick should be exactly 63 chars. Got %d", len(result))
	}
}