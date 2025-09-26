# Programs Package

The `programs` package provides program management functionality for the Master Thesis Operator Station. It allows remote launching, monitoring, and termination of Windows applications commonly used in flight simulation research.

## Overview

This package manages flight simulation applications by providing:
- Remote program launching capabilities
- Real-time program status monitoring
- Controlled program termination
- Process state management with permissions

## Key Features

### Program Management
- **Launch Programs**: Start applications remotely via HTTP API
- **Kill Programs**: Terminate running applications (with configurable permissions)
- **Status Monitoring**: Real-time monitoring of program running state
- **Process Validation**: Uses Windows `tasklist` for accurate process detection

### Supported Applications
The package is configured to manage these flight simulation applications:
- **FS2FF**: Flight data relay application (can be killed)
- **SkyDolly**: Flight recording application (protected from termination)
- **FS-FlightControl**: Flight control software (protected from termination)

### Safety Features
- **Kill Permissions**: Configurable protection for critical applications
- **Process Verification**: Double-checks process status using Windows system calls
- **State Synchronization**: Automatic monitoring to keep state accurate

## Architecture

### Core Components

**`programs.go`**
- Program initialization and configuration
- Process monitoring with 5-second intervals
- Windows process detection using `tasklist`

**`types.go`**
- Data structures for program definitions and state
- `Program`: Defines application metadata and permissions
- `ProgramState`: Tracks runtime state and process references

**`handlers.go`**
- HTTP handlers for REST API endpoints
- Event logging integration for audit trails
- Error handling and response management

## API Endpoints

### GET `/programs`
Returns list of all available programs.

**Response:**
```json
["FS2FF", "SkyDolly", "FS-FlightControl"]
```

### GET `/status?name=<program_name>`
Get current status of a specific program.

**Response:**
```json
{
  "running": true,
  "canKill": true
}
```

### POST `/launch?name=<program_name>`
Launch a program remotely.

**Success Response:** `200 OK`
**Error Responses:** 
- `404`: Program not found
- `409`: Program already running
- `500`: Launch failed

### POST `/kill?name=<program_name>`
Terminate a running program (if permitted).

**Success Response:** `200 OK`
**Error Responses:**
- `404`: Program not found or not running
- `500`: Termination failed

## Configuration

Programs are configured in the `Init()` function:

```go
programs["ProgramName"] = Program{
    Name:    "executable.exe",           // Process name for detection
    Path:    "C:\\Path\\To\\executable.exe", // Full path for launching
    CanKill: true,                       // Whether termination is allowed
}
```

## Event Integration

The package integrates with the `events` system to log:
- Program launches
- Program terminations
- Process state changes

Events are automatically logged with timestamps for audit purposes.

## Thread Safety

- Uses mutex locks for concurrent access to program states
- Background goroutine monitors process status every 5 seconds
- Atomic operations for state updates

## Platform Requirements

- **Windows**: Uses Windows-specific commands (`tasklist`, `taskkill`)
- **Administrative Access**: May require elevated privileges for some operations
- **Process Permissions**: Target applications must be accessible to the operator

## Usage Example

```go
// Initialize the programs package
programs.Init()

// Set up HTTP handlers
programs.SetupHandlers()

// The package will now handle HTTP requests for program management
```

## Error Handling

The package includes comprehensive error handling for:
- Invalid program names
- Process not found scenarios
- Permission denied situations
- Network/system failures

All errors are properly returned to clients with appropriate HTTP status codes.

## Security Considerations

- Kill permissions prevent accidental termination of critical applications
- Process validation prevents manipulation of unrelated processes
- Event logging provides audit trail for all operations
- Local-only operation intended (no authentication implemented)
