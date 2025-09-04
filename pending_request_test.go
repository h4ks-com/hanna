package main

import (
	"testing"
	"time"
)

func TestPendingRequestManagement(t *testing.T) {
	client := NewIRCClient()
	
	// Test creating a pending request
	req := client.createPendingRequest("list", "")
	if req == nil {
		t.Fatal("createPendingRequest returned nil")
	}
	if req.Type != "list" {
		t.Errorf("Expected type 'list', got '%s'", req.Type)
	}
	if req.Complete {
		t.Error("New request should not be complete")
	}
	
	// Test getting the pending request
	retrieved := client.getPendingRequest(req.ID)
	if retrieved == nil {
		t.Fatal("getPendingRequest returned nil")
	}
	if retrieved.ID != req.ID {
		t.Errorf("Retrieved request ID mismatch: expected %s, got %s", req.ID, retrieved.ID)
	}
	
	// Test finding request by type
	found := client.findPendingRequestByType("list")
	if found == nil {
		t.Fatal("findPendingRequestByType returned nil")
	}
	if found.ID != req.ID {
		t.Errorf("Found request ID mismatch: expected %s, got %s", req.ID, found.ID)
	}
	
	// Test completing the request
	client.completePendingRequest(req.ID)
	if !req.Complete {
		t.Error("Request should be complete after completePendingRequest")
	}
	
	// Test that completed request is no longer found by type
	found = client.findPendingRequestByType("list")
	if found != nil {
		t.Error("findPendingRequestByType should not return completed requests")
	}
}

func TestWhoisRequestManagement(t *testing.T) {
	client := NewIRCClient()
	
	// Test creating a WHOIS request
	req := client.createPendingRequest("whois", "testnick")
	if req == nil {
		t.Fatal("createPendingRequest returned nil")
	}
	if req.Type != "whois" {
		t.Errorf("Expected type 'whois', got '%s'", req.Type)
	}
	if req.Target != "testnick" {
		t.Errorf("Expected target 'testnick', got '%s'", req.Target)
	}
	
	// Test finding WHOIS request by nick
	found := client.findPendingWhoisRequest("testnick")
	if found == nil {
		t.Fatal("findPendingWhoisRequest returned nil")
	}
	if found.ID != req.ID {
		t.Errorf("Found request ID mismatch: expected %s, got %s", req.ID, found.ID)
	}
	
	// Test case-insensitive nick matching
	found = client.findPendingWhoisRequest("TESTNICK")
	if found == nil {
		t.Fatal("findPendingWhoisRequest should be case-insensitive")
	}
	
	// Test with different nick
	found = client.findPendingWhoisRequest("othernick")
	if found != nil {
		t.Error("findPendingWhoisRequest should not find different nick")
	}
}

func TestRequestTimeout(t *testing.T) {
	client := NewIRCClient()
	
	// Create a request with very short timeout for testing
	req := client.createPendingRequest("test", "")
	
	// Simulate timeout by waiting and checking if request is cleaned up
	// Note: In real implementation, cleanup happens after 30 seconds
	// For testing, we'll just verify the structure is correct
	
	if req.StartTime.IsZero() {
		t.Error("Request start time should be set")
	}
	
	if req.done == nil {
		t.Error("Request done channel should be initialized")
	}
	
	// Test GetRequestResult with immediate timeout
	result, err := client.GetRequestResult(req.ID, 1*time.Millisecond)
	if err == nil {
		t.Error("GetRequestResult should timeout for incomplete request")
	}
	if result != req {
		t.Error("GetRequestResult should return the request even on timeout")
	}
}
