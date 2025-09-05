package irc

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestFloodProtectionChannelDetection(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	oldMaxLines := os.Getenv("MAX_LINES_BEFORE_PASTING")
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
		os.Setenv("MAX_LINES_BEFORE_PASTING", oldMaxLines)
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Set test env vars
	os.Setenv("FLOOD_PROTECTED_CHANNELS", "#test,#bots,#dev")
	os.Setenv("MAX_LINES_BEFORE_PASTING", "2")
	os.Setenv("PASTE_CURL_TEMPLATE", "echo")
	
	client := NewClient()
	
	// Test protected channels
	if !client.isFloodProtectedChannel("#test") {
		t.Error("Expected #test to be flood protected")
	}
	
	if !client.isFloodProtectedChannel("#bots") {
		t.Error("Expected #bots to be flood protected")
	}
	
	if !client.isFloodProtectedChannel("#dev") {
		t.Error("Expected #dev to be flood protected")
	}
	
	// Test case insensitive matching
	if !client.isFloodProtectedChannel("#TEST") {
		t.Error("Expected case insensitive matching for #TEST")
	}
	
	// Test non-protected channels
	if client.isFloodProtectedChannel("#general") {
		t.Error("Expected #general to NOT be flood protected")
	}
	
	if client.isFloodProtectedChannel("#random") {
		t.Error("Expected #random to NOT be flood protected")
	}
	
	// Verify configuration loaded correctly
	if client.maxLinesBeforePasting != 2 {
		t.Errorf("Expected maxLinesBeforePasting to be 2, got %d", client.maxLinesBeforePasting)
	}
	
	if client.pasteCurlTemplate != "echo" {
		t.Errorf("Expected pasteCurlTemplate to be 'echo', got %s", client.pasteCurlTemplate)
	}
	
	expectedChannels := []string{"#test", "#bots", "#dev"}
	if len(client.floodProtectedChannels) != len(expectedChannels) {
		t.Errorf("Expected %d protected channels, got %d", len(expectedChannels), len(client.floodProtectedChannels))
	}
}

func TestFloodProtectionEmptyChannels(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
	}()
	
	// Set empty channels
	os.Setenv("FLOOD_PROTECTED_CHANNELS", "")
	
	client := NewClient()
	
	// No channels should be protected
	if client.isFloodProtectedChannel("#test") {
		t.Error("Expected no channels to be flood protected when FLOOD_PROTECTED_CHANNELS is empty")
	}
	
	if len(client.floodProtectedChannels) != 0 {
		t.Errorf("Expected 0 protected channels, got %d", len(client.floodProtectedChannels))
	}
}

func TestFloodProtectionDefaults(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	oldMaxLines := os.Getenv("MAX_LINES_BEFORE_PASTING")
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
		os.Setenv("MAX_LINES_BEFORE_PASTING", oldMaxLines)
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Clear env vars to test defaults
	os.Unsetenv("FLOOD_PROTECTED_CHANNELS")
	os.Unsetenv("MAX_LINES_BEFORE_PASTING")
	os.Unsetenv("PASTE_CURL_TEMPLATE")
	
	client := NewClient()
	
	// Test defaults
	if client.maxLinesBeforePasting != 3 {
		t.Errorf("Expected default maxLinesBeforePasting to be 3, got %d", client.maxLinesBeforePasting)
	}
	
	expectedTemplate := "curl -s -F \"file=@{{filename}}\" https://ix.io"
	if client.pasteCurlTemplate != expectedTemplate {
		t.Errorf("Expected default pasteCurlTemplate to be %s, got %s", expectedTemplate, client.pasteCurlTemplate)
	}
	
	if len(client.floodProtectedChannels) != 0 {
		t.Errorf("Expected 0 protected channels by default, got %d", len(client.floodProtectedChannels))
	}
}

func TestCreatePasteTemplate(t *testing.T) {
	// Save original env
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Set test curl template that just returns a test URL
	os.Setenv("PASTE_CURL_TEMPLATE", "echo http://test-paste-url")
	
	client := NewClient()
	
	testContent := "line1\nline2\nline3"
	
	// This should create a paste and return the "URL" (echo output)
	url, err := client.createPaste(testContent)
	if err != nil {
		t.Fatalf("Failed to create paste: %v", err)
	}
	
	// Since we're using echo, we should get back some output (the filename path)
	if url == "" {
		t.Error("Expected non-empty URL from paste creation")
	}
	
	// URL should not contain the template placeholder
	if strings.Contains(url, "{{filename}}") {
		t.Error("URL should not contain template placeholder")
	}
}

func TestCreatePasteInvalidTemplate(t *testing.T) {
	// Save original env
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Set invalid empty template (just whitespace)
	os.Setenv("PASTE_CURL_TEMPLATE", "   ")
	
	client := NewClient()
	
	testContent := "test content"
	
	_, err := client.createPaste(testContent)
	if err == nil {
		t.Error("Expected error with invalid curl template")
	}
	
	if !strings.Contains(err.Error(), "invalid curl template") {
		t.Errorf("Expected 'invalid curl template' error, got: %v", err)
	}
}

func TestCreatePasteFailingCommand(t *testing.T) {
	// Save original env
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Set a command that will fail
	os.Setenv("PASTE_CURL_TEMPLATE", "false")
	
	client := NewClient()
	
	testContent := "test content"
	
	_, err := client.createPaste(testContent)
	if err == nil {
		t.Error("Expected error with failing curl command")
	}
	
	if !strings.Contains(err.Error(), "curl command failed") {
		t.Errorf("Expected 'curl command failed' error, got: %v", err)
	}
}

func TestPrivmsgFloodProtectionTriggered(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	oldMaxLines := os.Getenv("MAX_LINES_BEFORE_PASTING")
	oldTemplate := os.Getenv("PASTE_CURL_TEMPLATE")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
		os.Setenv("MAX_LINES_BEFORE_PASTING", oldMaxLines)
		os.Setenv("PASTE_CURL_TEMPLATE", oldTemplate)
	}()
	
	// Set test env vars
	os.Setenv("FLOOD_PROTECTED_CHANNELS", "#test")
	os.Setenv("MAX_LINES_BEFORE_PASTING", "2")
	os.Setenv("PASTE_CURL_TEMPLATE", "echo http://test-paste-url")
	
	client := NewClient()
	
	// Mock the raw function to capture output
	var sentMessages []string
	client.testRawCapture = func(s string) {
		sentMessages = append(sentMessages, s)
	}
	
	// Test message with more lines than threshold
	multilineMessage := "line1\nline2\nline3\nline4\nline5"
	client.Privmsg("#test", multilineMessage)
	
	// Should send first 2 lines + paste URL message
	expectedMessages := 3 // 2 content lines + 1 paste URL line
	if len(sentMessages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(sentMessages))
	}
	
	// Check that we got the first two lines
	if len(sentMessages) >= 1 && !strings.Contains(sentMessages[0], "line1") {
		t.Errorf("Expected first message to contain 'line1', got: %s", sentMessages[0])
	}
	
	if len(sentMessages) >= 2 && !strings.Contains(sentMessages[1], "line2") {
		t.Errorf("Expected second message to contain 'line2', got: %s", sentMessages[1])
	}
	
	// Check that we got the paste URL
	if len(sentMessages) >= 3 && !strings.Contains(sentMessages[2], "http://test-paste-url") {
		t.Errorf("Expected paste URL message, got: %s", sentMessages[2])
	}
}

func TestPrivmsgNoFloodProtectionUnprotectedChannel(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	oldMaxLines := os.Getenv("MAX_LINES_BEFORE_PASTING")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
		os.Setenv("MAX_LINES_BEFORE_PASTING", oldMaxLines)
	}()
	
	// Set test env vars
	os.Setenv("FLOOD_PROTECTED_CHANNELS", "#test")
	os.Setenv("MAX_LINES_BEFORE_PASTING", "2")
	
	client := NewClient()
	
	// Mock the raw function to capture output
	var sentMessages []string
	client.testRawCapture = func(s string) {
		sentMessages = append(sentMessages, s)
	}
	
	// Test message to unprotected channel
	multilineMessage := "line1\nline2\nline3\nline4\nline5"
	client.Privmsg("#general", multilineMessage)
	
	// Should send all lines normally (no flood protection)
	expectedMessages := 5
	if len(sentMessages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(sentMessages))
	}
	
	// Check that all lines were sent
	for i := 0; i < 5; i++ {
		expectedContent := fmt.Sprintf("line%d", i+1)
		if !strings.Contains(sentMessages[i], expectedContent) {
			t.Errorf("Expected message %d to contain '%s', got: %s", i, expectedContent, sentMessages[i])
		}
	}
}

func TestPrivmsgNoFloodProtectionBelowThreshold(t *testing.T) {
	// Save original env
	oldChannels := os.Getenv("FLOOD_PROTECTED_CHANNELS")
	oldMaxLines := os.Getenv("MAX_LINES_BEFORE_PASTING")
	
	defer func() {
		os.Setenv("FLOOD_PROTECTED_CHANNELS", oldChannels)
		os.Setenv("MAX_LINES_BEFORE_PASTING", oldMaxLines)
	}()
	
	// Set test env vars
	os.Setenv("FLOOD_PROTECTED_CHANNELS", "#test")
	os.Setenv("MAX_LINES_BEFORE_PASTING", "5")
	
	client := NewClient()
	
	// Mock the raw function to capture output
	var sentMessages []string
	client.testRawCapture = func(s string) {
		sentMessages = append(sentMessages, s)
	}
	
	// Test message with fewer lines than threshold
	shortMessage := "line1\nline2\nline3"
	client.Privmsg("#test", shortMessage)
	
	// Should send all lines normally (below threshold)
	expectedMessages := 3
	if len(sentMessages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(sentMessages))
	}
}