# Data Analysis Module

The `data_analysis` package provides comprehensive flight data visualization and analysis capabilities for the Master Thesis Operator Station. It processes SQLite database files containing flight simulation data and generates interactive visualizations for research analysis.

## Overview

This module serves as a powerful analysis tool that:
- Processes SQLite databases from flight simulators
- Generates interactive visualizations of flight data
- Provides time-based filtering and annotation capabilities
- Supports multiple aircraft per flight analysis
- Offers real-time data exploration tools

## Key Features

### Database Support
- **Multiple Formats**: Supports `.sdlog`, `.sqlite`, and `.db` files
- **Flight Detection**: Automatically discovers flights in uploaded databases
- **Multi-Aircraft**: Handles multiple aircraft per flight session
- **Schema Validation**: Verifies required database structure

### Interactive Visualizations
- **Altitude Graphs**: Time-series altitude data with multiple aircraft support
- **GPS Mapping**: Interactive maps showing flight paths and position data
- **Airspeed Charts**: Velocity analysis with calculated airspeed from components
- **Real-time Updates**: Dynamic chart updates based on user controls

### Analysis Tools
- **Time Range Filtering**: Slider-based time range selection
- **Marker System**: Add custom annotations at specific time points
- **Preview Mode**: Live preview marker for real-time data exploration
- **Data Export**: Structured data access for further analysis

## Architecture

### Core Components

**`data_analysis.go`**
- Database connection management
- SQLite data processing and extraction
- HTTP handlers for file upload and data retrieval
- Flight data aggregation and calculation

**`types.go`**
- Data structure definitions for flights, aircraft, and position data
- Type safety for database records and API responses

**`page.templ`**
- HTML template for the main analysis interface
- Plotly.js integration for interactive visualizations
- Responsive design with tabbed interface

**`script.templ`**
- JavaScript functionality for interactive features
- Real-time chart updates and user interaction handling
- Data processing and visualization logic

## Data Structures

### Flight
```go
type Flight struct {
    ID           int    `json:"id"`
    Title        string `json:"title"`
    FlightNumber string `json:"flight_number"`
    StartTime    string `json:"start_time"`
    EndTime      string `json:"end_time"`
}
```

### Aircraft
```go
type Aircraft struct {
    ID           int    `json:"id"`
    FlightID     int    `json:"flight_id"`
    SeqNr        int    `json:"seq_nr"`
    Type         string `json:"type"`
    TailNumber   string `json:"tail_number"`
    Airline      string `json:"airline"`
}
```

### PositionPoint
```go
type PositionPoint struct {
    Timestamp         int64   `json:"timestamp"`
    TimestampSeconds  float64 `json:"timestamp_seconds"`
    Altitude          float64 `json:"altitude"`
    Latitude          float64 `json:"latitude"`
    Longitude         float64 `json:"longitude"`
    IndicatedAltitude float64 `json:"indicated_altitude"`
    PressureAltitude  float64 `json:"pressure_altitude"`
    Airspeed          float64 `json:"airspeed"`
}
```

### FlightData
```go
type FlightData struct {
    Flight       *Flight                       `json:"flight"`
    PositionData map[string][]PositionPoint   `json:"position_data"`
    EngineData   map[string][]EnginePoint     `json:"engine_data"`
}
```

## Database Schema

The module expects SQLite databases with the following structure:

### Required Tables

**`flight`**
- `id`: Primary key
- `title`: Flight description
- `flight_number`: Flight identifier
- `start_zulu_sim_time`: Flight start timestamp
- `end_zulu_sim_time`: Flight end timestamp

**`aircraft`**
- `id`: Primary key
- `flight_id`: Foreign key to flight table
- `seq_nr`: Aircraft sequence number
- `type`: Aircraft type/model
- `tail_number`: Aircraft registration
- `airline`: Airline identifier

**`position`**
- `aircraft_id`: Foreign key to aircraft table
- `timestamp`: Position timestamp (milliseconds)
- `altitude`: Altitude in meters
- `latitude`: Latitude in decimal degrees
- `longitude`: Longitude in decimal degrees
- `indicated_altitude`: Indicated altitude
- `pressure_altitude`: Pressure altitude

**`attitude`** (Optional)
- `aircraft_id`: Foreign key to aircraft table
- `timestamp`: Attitude timestamp
- `velocity_x`: X-axis velocity component
- `velocity_y`: Y-axis velocity component
- `velocity_z`: Z-axis velocity component

**`engine`** (Optional)
- `aircraft_id`: Foreign key to aircraft table
- `timestamp`: Engine data timestamp
- `throttle_lever_position1-4`: Throttle positions

## API Endpoints

### GET `/data-analysis`
Serve the main analysis interface.

**Response:** HTML page with embedded visualization tools

### POST `/data-analysis/upload`
Upload and process a SQLite database file.

**Request:** Multipart form with database file
**Response:**
```json
{
  "status": "success",
  "message": "Connected to database: flight_data.sdlog",
  "dbId": "uploaded_20250603_095605_flight_data.sdlog",
  "flights": [
    {
      "id": 1,
      "title": "Test Flight",
      "flight_number": "FL001",
      "start_time": "2025-06-03T09:00:00Z",
      "end_time": "2025-06-03T10:30:00Z"
    }
  ]
}
```

### GET `/data-analysis/flights?dbId=<id>`
Retrieve flights from an uploaded database.

**Response:**
```json
[
  {
    "id": 1,
    "title": "Test Flight",
    "flight_number": "FL001",
    "start_time": "2025-06-03T09:00:00Z",
    "end_time": "2025-06-03T10:30:00Z"
  }
]
```

### GET `/data-analysis/flight-data?dbId=<id>&flightId=<id>`
Retrieve complete flight data for analysis.

**Response:**
```json
{
  "flight": {
    "id": 1,
    "title": "Test Flight",
    "flight_number": "FL001"
  },
  "position_data": {
    "Cessna 172 (N12345)": [
      {
        "timestamp": 1717401600000,
        "timestamp_seconds": 0.0,
        "altitude": 152.4,
        "latitude": 54.9275,
        "longitude": -1.8342,
        "airspeed": 65.8
      }
    ]
  },
  "engine_data": {
    "Cessna 172 (N12345)": [
      {
        "timestamp": 1717401600000,
        "timestamp_seconds": 0.0,
        "throttle_position1": 0.75
      }
    ]
  }
}
```

### GET `/data-analysis/api/health`
Health check endpoint.

**Response:**
```json
{
  "status": "ok"
}
```

## Visualization Features

### Altitude Visualization
- Time-series plots of aircraft altitude
- Multiple aircraft overlay capability
- Altitude markers and annotations
- Time range filtering

### GPS Mapping
- Interactive maps using OpenStreetMap
- Flight path visualization
- Start/end point markers
- Automatic zoom and centering
- Real-time position tracking

### Airspeed Analysis
- Calculated airspeed from velocity components
- Time-based airspeed profiles
- Multi-aircraft comparison
- Performance analysis tools

## User Interface

### Upload Interface
- Drag-and-drop file upload
- File format validation
- Real-time upload progress
- Database connection verification

### Flight Selection
- Dropdown menu with flight details
- Flight metadata display
- Quick flight switching
- Data loading indicators

### Time Controls
- Dual-slider time range selection
- Real-time range updates
- Marker positioning slider
- Preview mode toggle

### Marker System
- Custom annotation placement
- Time-based marker positioning
- Label customization
- Marker management table

### Tabbed Visualizations
- Altitude, GPS, and Airspeed tabs
- Seamless tab switching
- Consistent data filtering
- Real-time updates across all views

## Data Processing

### Airspeed Calculation
```go
func calculateMagnitude(x, y, z float64) float64 {
    return sqrt(x*x + y*y + z*z)
}
```

### Time Synchronization
- Normalizes timestamps to seconds from flight start
- Matches attitude data to position data by timestamp
- Handles missing or sparse data gracefully

### Multi-Aircraft Handling
- Processes each aircraft independently
- Maintains aircraft identification in visualizations
- Supports different aircraft types in same flight

## Usage Workflow

1. **Database Upload**
   - Navigate to `/data-analysis`
   - Upload SQLite database file
   - Verify successful connection

2. **Flight Selection**
   - Choose flight from dropdown menu
   - Click "Load Flight Data"
   - Wait for data processing

3. **Data Exploration**
   - Use time range sliders to filter data
   - Switch between visualization tabs
   - Add markers for important events

4. **Analysis**
   - Examine altitude profiles
   - Analyze flight paths on map
   - Study airspeed characteristics
   - Export findings for further research

## Performance Optimization

### Database Queries
- Efficient SQL queries with proper indexing
- Minimal data transfer for large datasets
- Chunked data processing for memory efficiency

### Frontend Performance
- Client-side data caching
- Optimized chart rendering with Plotly.js
- Debounced user input handling
- Progressive data loading

### Memory Management
- Temporary file cleanup
- Database connection pooling
- Garbage collection for large datasets

## Error Handling

### File Upload
- Format validation before processing
- Size limit enforcement
- Corruption detection
- User-friendly error messages

### Database Processing
- Schema validation
- Missing table detection
- Data integrity checks
- Graceful degradation for optional tables

### Visualization
- Missing data handling
- Invalid coordinate filtering
- Chart rendering error recovery
- User interaction error prevention

## Configuration

### File Storage
- Temporary uploads stored in `temp_uploads/`
- Automatic cleanup of old files
- Configurable storage limits

### Database Connections
- Connection pooling for efficiency
- Automatic connection cleanup
- Timeout handling

### Visualization Settings
- Configurable chart dimensions
- Default time ranges
- Color schemes for multiple aircraft
- Map tile server configuration

## Integration

### Events System
- Logs analysis session events
- Tracks user interactions
- Records data processing milestones

### Main Application
- Consistent routing patterns
- Shared error handling
- Common UI components

### External Dependencies
- Plotly.js for interactive charts
- OpenStreetMap for mapping
- Templ for HTML templating

This module provides researchers with powerful tools for analyzing flight simulation data, enabling detailed examination of aircraft performance, flight patterns, and operational characteristics through interactive visualizations and comprehensive data processing capabilities.
