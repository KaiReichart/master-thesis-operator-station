# Master Thesis Operator Station

> **WARNING**: 100% of this project has been vibe-coded. It works only for EXACTLY my use-case, may break at any time, and almost certainly contain grave security flaws.

A comprehensive web-based research platform for flight simulation studies, providing integrated tools for program management, GPS tracking, psychological testing, flight data analysis, and event logging. Built in Go with a modern web interface, this system supports aviation human factors research and flight simulation experiments.

## 🎯 Overview

The Master Thesis Operator Station is a modular research platform designed for aviation psychology and flight simulation research. It provides researchers with a unified interface to manage flight simulation software, track aircraft positions, conduct psychological assessments, analyze flight data, and maintain detailed audit logs of experimental sessions.

The platform features a comprehensive **overview page** that serves as a central navigation hub, providing easy access to all system modules through an intuitive card-based interface. From the overview, researchers can quickly navigate to the Program Manager for flight software control, the Mental Rotation Test for cognitive assessments, or the Data Analysis module for flight data visualization.

## 🏗️ Architecture

The system is built using a modular architecture with five core packages:

- **Programs**: Remote management of flight simulation applications
- **GPS**: Real-time position tracking and intelligent data relay
- **Mental Rotation**: Computerized spatial cognition testing
- **Data Analysis**: Interactive flight data visualization and analysis
- **Events**: Comprehensive audit logging and event management

## 📦 Core Modules

### 🖥️ Program Manager (`programs/`)
Remote control and monitoring of Windows flight simulation applications.

**Key Features:**
- Launch and terminate applications remotely
- Real-time status monitoring with 5-second intervals
- Configurable kill permissions for safety-critical applications
- Process validation using Windows system calls

**Supported Applications:**
- FS2FF (Flight data relay)
- SkyDolly (Flight recording)
- FS-FlightControl (Flight control software)

**API Endpoints:**
- `GET /programs` - List available programs
- `GET /status` - Check program status
- `POST /launch` - Start applications
- `POST /kill` - Terminate applications

### 🛰️ GPS Module (`gps/`)
Intelligent GPS data processing with proximity-based forwarding.

**Key Features:**
- UDP listener for FS2FF GPS broadcasts (port 49002)
- Real-time WebSocket position updates
- Distance-based automatic data forwarding
- Configurable target IP and distance thresholds
- Reference point: Currock Hill (54.9275°N, 1.8342°W)

**Data Processing:**
- Parses XGPS packet format from FS2FF
- Calculates distance using Haversine formula
- Converts altitude from feet to meters
- Validates GPS coordinates and timing

**API Endpoints:**
- `WebSocket /gps-ws` - Real-time position updates
- `POST /set-target-ip` - Configure forwarding destination
- `POST /set-distance-threshold` - Set proximity limits
- `POST /broadcast-toggle` - Manual forwarding control

### 🧠 Mental Rotation Test (`mental_rotation/`)
Psychological assessment tool for spatial cognitive abilities.

**Key Features:**
- Automated 3D object rotation tasks
- German language interface for research compliance
- Precise millisecond timing measurements
- Participant tracking and result storage
- Embedded image sets with automatic task generation

**Test Protocol:**
- Participant ID validation
- Sequential task presentation
- Binary choice responses (same/different objects)
- Automatic scoring based on filename conventions
- JSON result export for statistical analysis

**Image Convention:**
- Standard images: Same object in different orientations
- `_R.jpg` suffix: Reflected/mirrored objects (different shapes)

### 📊 Data Analysis Module (`data_analysis/`)
Interactive flight data visualization and analysis platform.

**Key Features:**
- SQLite database processing (`.sdlog`, `.sqlite`, `.db`)
- Multi-aircraft flight analysis support
- Interactive Plotly.js visualizations
- Time-based filtering and annotation
- Real-time data exploration tools

**Visualization Types:**
- **Altitude Graphs**: Time-series altitude profiles
- **GPS Mapping**: Interactive flight path visualization
- **Airspeed Charts**: Velocity analysis with calculated airspeeds

**Analysis Tools:**
- Dual-slider time range selection
- Custom marker annotation system
- Preview mode for real-time exploration
- Multi-aircraft overlay capabilities

### 📝 Events System (`events/`)
Comprehensive audit logging and event management.

**Key Features:**
- Real-time event logging to timestamped files
- In-memory storage for quick access (last 50 events)
- RESTful API for event retrieval and manual recording
- Thread-safe operations for concurrent access

**Event Types:**
- Program operations (launch, kill)
- Flight phases (started, ended, preparations)
- Failure management (started, recognised, recovery)
- Operator states (confused, back on track)
- GPS operations (configuration, state changes)

**Storage Format:**
```
logs/events_YYYY-MM-DD_HH-MM-SS.log
[YYYY-MM-DD HH:MM:SS] EVENT_TYPE: program_name
```

## 🚀 Quick Start

### Prerequisites
- **Operating System**: Windows (for flight simulation application compatibility)
- **Go Version**: 1.24 or later
- **Network**: Access to port 8080 and UDP port 49002
- **Flight Simulator**: FS2FF configured for GPS broadcasting

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd master-thesis-operator-station
   ```

2. **Build the application:**
   ```bash
   go mod tidy
   make build
   # or directly: go build
   ```

3. **Run the server:**
   ```bash
   ./master-thesis-operator-station
   # or: make run (for development with hot reload)
   ```

4. **Access the application:**
   - Overview page: `http://localhost:8080`
   - Program Manager: `http://localhost:8080/program-manager`
   - Flight Data Analysis: `http://localhost:8080/data-analysis`
   - Mental Rotation Test: `http://localhost:8080/mental-rotation`

### Configuration

Program paths are configured in `main.go`:

```go
programs["ProgramName"] = Program{
    Name:    "executable.exe",
    Path:    "C:\\Path\\To\\executable.exe",
    CanKill: true,  // Safety setting
}
```

## 🌐 Web Interface

### Overview Page (`/`)
- **Navigation Hub**: Central access point to all platform modules
- **Module Cards**: Interactive cards with descriptions and quick access
- **System Status**: Real-time platform information and health indicators
- **Quick Access**: Shortcuts to common research tasks

### Program Manager (`/program-manager`)
- **Program Status**: Real-time monitoring of flight simulation applications
- **GPS Position**: Live aircraft position with coordinates and distance calculations
- **Event Logging**: Manual event recording with categorized buttons
- **Configuration**: Target IP and distance threshold settings

### Flight Data Analysis
- **Database Upload**: Drag-and-drop SQLite file processing
- **Flight Selection**: Dropdown menu with flight metadata
- **Interactive Controls**: Time range sliders and marker annotation
- **Tabbed Visualizations**: Altitude, GPS maps, and airspeed charts

### Mental Rotation Testing
- **Participant Management**: ID-based session tracking
- **Task Presentation**: Full-screen 3D object display
- **Response Recording**: Precise timing and accuracy measurement
- **Results Export**: JSON format for statistical analysis

## 🔧 Development

### Build Commands
```bash
make run        # Development server with hot reload
make build      # Production build
make generate   # Generate templates and assets
make clean      # Clean build artifacts
templ generate  # Generate Go code from templ files
```

### Project Structure
```
├── main.go                 # Application entry point
├── overview.html           # Overview/landing page
├── program-manager.html    # Program manager interface
├── programs/              # Program management module
├── gps/                   # GPS tracking and relay
├── mental_rotation/       # Psychological testing
├── data_analysis/         # Flight data visualization
├── events/                # Event logging system
├── data/                  # Data storage directory
├── logs/                  # Event log files
└── temp_uploads/          # Temporary file storage
```

### Technology Stack
- **Backend**: Go 1.24 with native HTTP server
- **Frontend**: HTML5, CSS3, JavaScript (ES6+)
- **Visualization**: Plotly.js for interactive charts
- **Templating**: Templ for type-safe HTML generation
- **Database**: SQLite for data storage
- **WebSocket**: Real-time GPS data streaming
- **Build Tools**: Make, Go modules, Air (hot reload)

## 📡 API Reference

### RESTful Endpoints
```
# Web Interface
GET    /                            # Overview page
GET    /program-manager             # Program manager interface
GET    /mental-rotation             # Mental rotation test
GET    /data-analysis               # Data analysis interface

# Program Management
GET    /programs                    # List available programs
GET    /status?name=<program>       # Get program status
POST   /launch?name=<program>       # Launch program
POST   /kill?name=<program>         # Kill program

# Event Management
GET    /events                      # Get recent events
POST   /manual-event               # Record manual event

# GPS Configuration
POST   /set-target-ip              # Configure target IP
GET    /get-target-ip              # Get current target IP
POST   /broadcast-toggle           # Toggle GPS broadcasting
POST   /set-distance-threshold     # Set distance limit

# Data Analysis
POST   /data-analysis/upload       # Upload database
GET    /data-analysis/flights      # Get flight list
GET    /data-analysis/flight-data  # Get flight data

# Mental Rotation
GET    /mental-rotation/tasks      # Get test tasks
POST   /mental-rotation/submit     # Submit test result
GET    /mental-rotation/results    # Get all results
```

### WebSocket Endpoints
```
ws://localhost:8080/gps-ws         # Real-time GPS position updates
```

## 🔒 Security & Safety

### Program Management Safety
- Configurable kill permissions prevent accidental termination
- Process validation ensures only intended applications are managed
- Event logging provides complete audit trail

### Network Security
- Local-only operation by default (no external authentication)
- UDP port validation for GPS data
- IP address validation for target configuration

### Data Protection
- Local file storage with appropriate permissions
- No sensitive data transmission over unencrypted channels
- Participant data anonymization support

## 🎯 Research Applications

### Aviation Human Factors
- Monitor pilot performance during simulated flights
- Track aircraft position relative to study areas
- Record operator interventions and decision points
- Analyze flight patterns and performance metrics

### Cognitive Assessment
- Measure spatial rotation abilities
- Correlate cognitive performance with flight performance
- Track learning curves and skill development
- Support experimental psychology research

### Flight Training Analysis
- Record training session events and milestones
- Analyze flight data for instruction purposes
- Monitor student progress and performance
- Provide objective performance metrics

## 🔧 Troubleshooting

### Common Issues

**GPS Not Receiving Data:**
- Verify FS2FF is broadcasting on port 49002
- Check Windows firewall settings
- Ensure UDP port is not blocked

**Program Launch Failures:**
- Verify program paths in configuration
- Check file permissions and execution rights
- Ensure applications are not already running

**Database Upload Issues:**
- Validate SQLite file format and schema
- Check file size limits and available disk space
- Verify required database tables exist

### Log Files
- **Event Logs**: `logs/events_YYYY-MM-DD_HH-MM-SS.log`
- **Application Logs**: Console output during development
- **Error Logs**: Check terminal/console for runtime errors

## 🤝 Contributing

### Development Setup
1. Install Go 1.24+
2. Install development tools: `go install github.com/air-verse/air@latest`
3. Run development server: `make run`
4. Make changes and test locally

### Code Standards
- Follow Go best practices and conventions
- Include comprehensive error handling
- Add unit tests for new functionality
- Document API changes and new features

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Update documentation as needed
5. Submit pull request with clear description

## 📄 License

This project is developed for academic research purposes. Please refer to the license file for specific terms and conditions.

## 🙏 Acknowledgments

Developed for master's thesis research in aviation psychology and human factors. Special thanks to the flight simulation community for tools and protocols that make this research possible.

---

**Version**: 1.1.0  
**Last Updated**: June 2025  
**Maintainer**: Research Team  
**Platform**: Windows 10/11, Go 1.24+

## 📋 Recent Updates

### Version 1.2.0 (June 2025)
- **MIGRATED**: Converted HTML templates to type-safe Templ templates
- **IMPROVED**: Better template organization with partials and layouts
- **ENHANCED**: Type safety for template parameters and rendering
- **UPDATED**: All packages (programs, gps, events) now use Templ templating

### Version 1.1.0 (June 2025)
- **NEW**: Added comprehensive overview page at root URL (`/`)
- **REORGANIZED**: Moved program manager functionality to `/program-manager`
- **ENHANCED**: Improved navigation with intuitive card-based interface
- **ADDED**: System status indicators and quick access shortcuts
- **UPDATED**: Documentation to reflect new routing structure
