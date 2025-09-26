# Events Package

The `events` package provides comprehensive event logging and audit trail functionality for the Master Thesis Operator Station. It captures, stores, and manages operational events across all system modules.

## Overview

This package serves as the central event management system, providing:
- Real-time event logging to files
- In-memory event storage for quick access
- RESTful API for event retrieval and manual event recording
- Thread-safe operations for concurrent access

## Key Features

### Event Logging
- **File-based Logging**: Automatic timestamped log files in `/logs` directory
- **In-memory Storage**: Fast access to recent events (last 50)
- **Structured Data**: JSON-serializable event objects
- **Thread Safety**: Mutex-protected operations for concurrent access

### Event Types
The system supports various event types including:
- **Program Events**: `launch`, `kill`
- **Flight Operations**: `flight_started`, `flight_ended`
- **Failure Management**: `failure_started`, `failure_recognised`, `back_on_track`
- **Operator State**: `confused`
- **GPS Operations**: `sending_toggled`, `target_ip_set`, `distance_threshold_updated`
- **Custom Events**: User-defined events through manual logging

### Audit Trail
- Immutable event records with precise timestamps
- Persistent storage survives application restarts
- Chronological ordering for event sequence analysis
- Integration with all system modules for comprehensive coverage

## Architecture

### Core Components

**`events.go`**
- Event storage and file management
- Thread-safe event logging with mutex protection
- Automatic log file creation with timestamps
- In-memory event array for performance

**`types.go`**
- Event data structure definition
- Timestamp management
- Type safety for event properties

**`handlers.go`**
- REST API endpoints for event operations
- Manual event recording functionality
- Event retrieval with pagination (last 50 events)

## Data Structure

### Event Type
```go
type Event struct {
    Type      string    `json:"type"`      // Event type identifier
    Program   string    `json:"program"`   // Associated program/module
    Timestamp time.Time `json:"timestamp"` // Precise occurrence time
}
```

## API Endpoints

### GET `/events`
Retrieve recent events (last 50).

**Response:**
```json
[
  {
    "type": "launch",
    "program": "FS2FF",
    "timestamp": "2025-06-03T10:30:45.123Z"
  },
  {
    "type": "flight_started",
    "program": "Operator",
    "timestamp": "2025-06-03T10:35:12.456Z"
  }
]
```

### POST `/manual-event`
Record a manual event.

**Request Body:**
```json
{
  "type": "failure_recognised",
  "program": "Operator"
}
```

**Success Response:** `200 OK`

## File Logging Format

Log files are created in the `logs/` directory with format:
```
logs/events_YYYY-MM-DD_HH-MM-SS.log
```

**Log Entry Format:**
```
[YYYY-MM-DD HH:MM:SS] EVENT_TYPE: program_name
```

**Example:**
```
=== Event Log Started at 2025-06-03 10:30:00 ===
[2025-06-03 10:30:45] LAUNCH: FS2FF
[2025-06-03 10:35:12] FLIGHT_STARTED: Operator
[2025-06-03 10:40:23] FAILURE_STARTED: Operator
```

## Event Types Reference

### Program Management
- `launch`: Application started
- `kill`: Application terminated

### Flight Operations
- `flight_started`: Flight session begins
- `flight_ended`: Flight session ends
- `preparations_started`: Pre-flight preparations begin
- `preparations_finished`: Pre-flight preparations complete

### Failure Management
- `failure_started`: System or procedural failure detected
- `failure_recognised`: Operator acknowledges failure
- `back_on_track`: Recovery from failure state

### Operator State
- `confused`: Operator reports confusion or uncertainty

### GPS Operations
- `sending_toggled`: GPS broadcast state changed
- `target_ip_set`: Target IP address configured
- `distance_threshold_updated`: Distance threshold modified
- `reached_target`: GPS position within target range

## Usage Examples

### Recording Events Programmatically
```go
// Create an event
event := events.Event{
    Type:      "launch",
    Program:   "FS2FF",
    Timestamp: time.Now(),
}

// Log the event
events.LogEvent(event)
```

### Manual Event Recording via API
```javascript
// Record a manual event
fetch('/manual-event', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        type: 'failure_recognised',
        program: 'Operator'
    })
});
```

### Retrieving Events
```javascript
// Get recent events
fetch('/events')
    .then(response => response.json())
    .then(events => {
        console.log('Recent events:', events);
    });
```

## Integration

The events package is integrated across all system modules:

- **Programs**: Automatic logging of launch/kill operations
- **GPS**: Logging of configuration changes and state transitions
- **Data Analysis**: Event recording for analysis sessions
- **Mental Rotation**: Logging of test completion events
- **Frontend**: Manual event recording via user interface

## Thread Safety

All operations are thread-safe through:
- Mutex protection for shared data structures
- Atomic append operations for event arrays
- Synchronized file I/O operations

## File Management

- **Automatic Creation**: Log files created on application startup
- **Unique Naming**: Timestamp-based naming prevents conflicts
- **Directory Structure**: Organized in dedicated `logs/` directory
- **Error Handling**: Graceful degradation if file operations fail

## Performance Considerations

- **In-Memory Caching**: Recent events stored in memory for fast access
- **Limited Retention**: In-memory storage limited to 50 events
- **Asynchronous Logging**: File operations don't block event recording
- **Efficient Serialization**: Minimal overhead for JSON operations

## Error Handling

The package includes robust error handling for:
- File system access issues
- JSON serialization errors
- Network request failures
- Invalid event data

Errors are logged appropriately and don't prevent continued operation.

## Configuration

The package initializes automatically with:
- Log directory creation
- Initial log file setup
- Memory structure allocation
- Error recovery mechanisms

No external configuration required for basic operation.
