# GPS Package

The `gps` package provides real-time GPS position tracking, UDP data relay, and distance-based GPS broadcasting functionality for flight simulation research. It listens for FS2FF GPS broadcasts, processes position data, and conditionally forwards data based on proximity to target locations.

## Overview

This package serves as a GPS data hub that:
- Receives UDP broadcasts from FS2FF flight simulator GPS data
- Processes and validates GPS position information
- Provides real-time position updates via WebSocket
- Implements distance-based GPS data forwarding
- Manages target IP configuration for data relay

## Key Features

### GPS Data Processing
- **UDP Listener**: Receives FS2FF XGPS packets on port 49002
- **Data Parsing**: Processes comma-separated GPS coordinate data
- **Real-time Updates**: Broadcasts position updates via WebSocket
- **Distance Calculation**: Computes distance to reference points (Currock Hill)

### Intelligent Relay System
- **Proximity-based Forwarding**: Only forwards GPS data when within specified distance
- **Configurable Thresholds**: Adjustable distance limits (default: 10 nautical miles)
- **Target IP Management**: Dynamic configuration of destination IP addresses
- **Automatic State Management**: Toggles forwarding based on position

### Reference Location
- **Currock Hill Coordinates**: 54.9275°N, 1.8342°W
- **Distance Monitoring**: Continuous calculation of distance from reference point
- **Threshold Management**: Configurable maximum distance for data forwarding

## Architecture

### Core Components

**`gps.go`**
- UDP listener for FS2FF GPS broadcasts
- GPS data processing and validation
- WebSocket connection management
- Distance calculation and threshold management
- Automatic GPS forwarding logic

**`types.go`**
- Data structures for GPS positions and raw GPS data
- Type definitions for position coordinates and metadata

**`handlers.go`**
- REST API endpoints for GPS configuration
- WebSocket handler for real-time position updates
- Target IP management and broadcasting control

**`helpers.go`**
- GPS packet parsing utilities
- Distance calculation functions (Haversine formula)
- Coordinate conversion and validation

## Data Structures

### GPSPosition
```go
type GPSPosition struct {
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    Altitude  float64   `json:"altitude"`   // Converted to meters
    Timestamp time.Time `json:"timestamp"`
}
```

### GPSData (Raw)
```go
type GPSData struct {
    Latitude      float32  // Decimal degrees
    Longitude     float32  // Decimal degrees
    AltitudeMSL   float32  // Mean Sea Level altitude (feet)
    GroundSpeed   float32  // Ground speed
    TrueHeading   float32  // True heading (degrees)
    MagHeading    float32  // Magnetic heading (degrees)
    IAS           float32  // Indicated Airspeed
    TAS           float32  // True Airspeed
    VerticalSpeed float32  // Vertical speed
}
```

## UDP Data Format

The package expects XGPS packets with format:
```
XGPS[length],longitude,latitude,altitude,heading,speed
```

**Example:**
```
XGPS25,-1.834200,54.927500,152.3,090.5,125.2
```

## API Endpoints

### WebSocket `/gps-ws`
Real-time GPS position updates.

**Message Format:**
```json
{
  "latitude": 54.927500,
  "longitude": -1.834200,
  "altitude": 46.4,
  "timestamp": "2025-06-03T10:30:45.123Z"
}
```

### POST `/set-target-ip`
Configure target IP for GPS data forwarding.

**Request Body:**
```json
{
  "ip": "192.168.1.100"
}
```

### GET `/get-target-ip`
Retrieve current target IP configuration.

**Response:**
```json
{
  "ip": "192.168.1.100"
}
```

### POST `/broadcast-toggle`
Manual control of GPS broadcasting state.

**Request Body:**
```json
{
  "enabled": true
}
```

### POST `/set-distance-threshold`
Configure maximum distance for automatic forwarding.

**Request Body:**
```json
{
  "threshold": 15.0
}
```

## Distance Calculation

Uses the Haversine formula for great-circle distance calculation:

```go
func calculateDistanceNM(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 3440.065 // Earth's radius in nautical miles
    // ... Haversine calculation
    return R * c
}
```

## Automatic Forwarding Logic

GPS data is automatically forwarded when:
1. Current position is within configured distance threshold of Currock Hill
2. Target IP address is configured
3. GPS data is valid and recent

The system automatically:
- Calculates distance to reference point for each GPS update
- Toggles forwarding state based on proximity
- Logs state changes through the events system

## Event Integration

Generates events for:
- `sending_toggled`: When GPS forwarding state changes
- `target_ip_set`: When target IP is configured
- `distance_threshold_updated`: When threshold is modified

## Usage Examples

### WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/gps-ws');

ws.onmessage = function(event) {
    const position = JSON.parse(event.data);
    console.log(`Position: ${position.latitude}, ${position.longitude}`);
};
```

### Setting Target IP
```javascript
fetch('/set-target-ip', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        ip: '192.168.1.100'
    })
});
```

### Configuring Distance Threshold
```javascript
fetch('/set-distance-threshold', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        threshold: 15.0  // 15 nautical miles
    })
});
```

## GPS Data Flow

1. **UDP Reception**: FS2FF broadcasts XGPS packets on port 49002
2. **Packet Validation**: Checks for XGPS header and minimum length
3. **Data Parsing**: Extracts coordinates, altitude, and flight parameters
4. **Position Update**: Converts to standard GPSPosition format
5. **Distance Calculation**: Computes distance to Currock Hill reference
6. **Forwarding Decision**: Determines if data should be relayed
7. **WebSocket Broadcast**: Sends updates to connected clients
8. **UDP Forward**: Relays packets to target IP if within threshold

## Configuration

### Default Settings
- **UDP Port**: 49002 (standard FS2FF port)
- **Reference Point**: Currock Hill (54.9275°N, 1.8342°W)
- **Default Threshold**: 10.0 nautical miles
- **Default Target IP**: 192.168.178.152

### Coordinate System
- **Input Format**: Decimal degrees (FS2FF standard)
- **Altitude Units**: Converted from feet to meters
- **Distance Units**: Nautical miles for thresholds

## Error Handling

Robust error handling for:
- Invalid UDP packet formats
- GPS data parsing errors
- Network connectivity issues
- WebSocket connection failures
- Target IP validation

## Performance Considerations

- **Efficient UDP Processing**: Minimal overhead for packet processing
- **Real-time Updates**: Sub-second GPS position updates
- **WebSocket Management**: Automatic cleanup of disconnected clients
- **Distance Calculations**: Optimized Haversine formula implementation

## Thread Safety

All operations are thread-safe through:
- Mutex protection for shared GPS state
- Concurrent WebSocket client management
- Atomic operations for configuration updates

## Integration Requirements

- **FS2FF**: Must be configured to broadcast on port 49002
- **Network Access**: UDP port 49002 must be accessible
- **Target Network**: Destination IP must be reachable for forwarding
- **Events System**: Requires events package for logging

This package provides essential GPS functionality for flight simulation research, enabling intelligent data relay based on aircraft position relative to study areas.
