# Unit Tests

This directory contains all unit tests for the Hanna IRC Bot.

## Running Tests

### From the main project directory:
```bash
go test ./UnitTests/
```

### From the UnitTests directory:
```bash
cd UnitTests
go test -v
```

## Test Files

- **`channel_state_test.go`** - Tests for channel state tracking functionality
  - Channel user management
  - IRC mode parsing and application  
  - Command handling (JOIN, PART, QUIT, MODE, etc.)
  - API state endpoint output format

- **`main_test.go`** - Tests for core IRC functionality
  - Nick mention detection and pattern matching
  - Message filtering logic

## Test Coverage

The tests verify:

✅ **Channel State Management**
- Adding/removing users from channels
- Tracking user modes (operator, voice, halfop)
- Handling user nick changes across channels
- Proper cleanup on PART/QUIT/KICK events

✅ **IRC Mode Parsing**
- Parsing complex mode strings like `+oo-v+h nick1 nick2 nick3`
- Applying mode changes to user state
- Handling mode prefixes in NAMES replies (`@`, `+`, `%`)

✅ **API Output Format**
- Ensures `/api/state` returns the correct JSON structure
- Verifies `null` values for users without modes
- Tests serialization/deserialization

✅ **Nick Matching**
- Word boundary detection for bot mentions
- Case-insensitive matching
- Filtering mentions surrounded by ignore characters

## Notes

- Tests use manual state manipulation instead of actual IRC connections for reliability
- All tests are designed to run without network dependencies
- The `main.go` file is copied to this directory so tests have access to all functions
