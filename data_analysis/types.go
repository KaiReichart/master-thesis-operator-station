package data_analysis

// Flight represents a flight record from the database
type Flight struct {
	ID           int    `json:"id"`
	SourceID     int    `json:"source_id,omitempty"` // ID from original database for import tracking
	Title        string `json:"title"`
	FlightNumber string `json:"flight_number"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

// Aircraft represents an aircraft in a flight
type Aircraft struct {
	ID         int    `json:"id"`
	FlightID   int    `json:"flight_id"`
	SeqNr      int    `json:"seq_nr"`
	Type       string `json:"type"`
	TailNumber string `json:"tail_number"`
	Airline    string `json:"airline"`
}

// PositionPoint represents a single position data point
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

// EnginePoint represents a single engine data point
type EnginePoint struct {
	Timestamp         int64   `json:"timestamp"`
	TimestampSeconds  float64 `json:"timestamp_seconds"`
	ThrottlePosition1 float64 `json:"throttle_position1"`
	ThrottlePosition2 float64 `json:"throttle_position2"`
	ThrottlePosition3 float64 `json:"throttle_position3"`
	ThrottlePosition4 float64 `json:"throttle_position4"`
}

// FlightData represents all data for a flight
type FlightData struct {
	Flight       *Flight                    `json:"flight"`
	PositionData map[string][]PositionPoint `json:"position_data"`
	EngineData   map[string][]EnginePoint   `json:"engine_data"`
}

// Marker represents a user-defined marker on the timeline
type Marker struct {
	ID        int     `json:"id"`
	FlightID  int     `json:"flight_id"`
	Time      float64 `json:"time"`
	Label     string  `json:"label"`
	Type      string  `json:"type"` // "regular", "trim_start", "trim_end"
	CreatedAt string  `json:"created_at,omitempty"`
}

// VisualizationRequest represents a request for generating visualizations
type VisualizationRequest struct {
	FlightData  *FlightData `json:"flight_data"`
	StartTime   float64     `json:"start_time"`
	EndTime     float64     `json:"end_time"`
	Markers     []Marker    `json:"markers"`
	PreviewTime *float64    `json:"preview_time"`
}

// DatabaseInfo represents information about an available database
type DatabaseInfo struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ModTime     string `json:"mod_time"`
	FlightCount int    `json:"flight_count"`
}

// CSVFlightData represents flight data parsed from a CSV file
type CSVFlightData struct {
	Metadata CSVMetadata       `json:"metadata"`
	Headers  []string          `json:"headers"`
	Records  []CSVFlightRecord `json:"records"`
}

// CSVMetadata contains metadata about the CSV file
type CSVMetadata struct {
	Source       string `json:"source"`        // e.g., "FS-FlightControl"
	RecordedAt   string `json:"recorded_at"`   // Original recording timestamp
	FlightTitle  string `json:"flight_title"`  // User-provided or derived title
	AircraftType string `json:"aircraft_type"` // User-provided aircraft type
	TotalRecords int    `json:"total_records"`
}

// CSVFlightRecord represents a single data point from CSV
type CSVFlightRecord struct {
	// Time data
	Time             string  `csv:"Time"`
	TimestampSeconds float64 `json:"timestamp_seconds"`

	// Airspeed data
	AirspeedIndicated float64 `csv:"AirspeedIndicated (knots)"`
	AirspeedTrue      float64 `csv:"AirspeedTrue (knots)"`
	GroundSpeed       float64 `csv:"GroundSpeed (knots)"`

	// Altitude data
	Altitude        float64 `csv:"Altitude (feet)"`
	GroundElevation float64 `csv:"GroundElevation (meters)"`

	// Position data
	Latitude  float64 `csv:"Latitude (degrees)"`
	Longitude float64 `csv:"Longitude (degrees)"`

	// Attitude data
	BankAngle       float64 `csv:"BankAngle (degrees)"`
	PitchAngle      float64 `csv:"PitchAngle (degrees)"`
	HeadingMagnetic float64 `csv:"HeadingMagnetic (degrees)"`
	HeadingTrue     float64 `csv:"HeadingTrue (degrees)"`

	// Environmental data
	AmbientTemperature   float64 `csv:"AmbientTemperature (celsius)"`
	AmbientWindDirection float64 `csv:"AmbientWindDirection (degrees)"`
	AmbientWindVelocity  float64 `csv:"AmbientWindVelocity (knots)"`

	// Aircraft state
	FlapsHandlePosition float64 `csv:"FlapsHandlePosition"`
	FuelTotalQuantity   float64 `csv:"FuelTotalQuantity (gallons)"`
	GearDown            bool    `csv:"GearDown (bool)"`
	OnGround            bool    `csv:"OnGround (bool)"`

	// Flight dynamics
	GForce        float64 `csv:"GForce (gforce)"`
	VerticalSpeed float64 `csv:"VerticalSpeed (feet per minute)"`

	// Warnings and alerts
	OverspeedWarning bool `csv:"OverspeedWarning (bool)"`
	StallWarning     bool `csv:"StallWarning (bool)"`
}

// CSVImportOptions defines options for CSV import
type CSVImportOptions struct {
	FlightTitle  string `json:"flight_title"`
	AircraftType string `json:"aircraft_type"`
	SkipRows     int    `json:"skip_rows"` // Number of header rows to skip
}
