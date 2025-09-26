package data_analysis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate go tool templ generate

var (
	tempDir = "temp_uploads"
	
	// Currock Hill coordinates (shared from GPS module)
	currockHillLat = 54.9275
	currockHillLon = -1.8342
	targetDistanceNM = 9.0
)

func Init() {
	// Create temp directory for uploaded databases
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Failed to create temp directory: %v", err)
	}

	// Initialize the main database
	if err := InitMainDatabase(); err != nil {
		log.Fatalf("Failed to initialize main database: %v", err)
	}

	log.Println("Data Analysis module initialized")
}

func SetupHandlers() {
	http.HandleFunc("/data-analysis", serveDataAnalysisPage)
	http.HandleFunc("/data-analysis/upload", handleDatabaseUpload)
	http.HandleFunc("/data-analysis/flights", handleGetFlights)
	http.HandleFunc("/data-analysis/flight-data", handleGetFlightData)
	http.HandleFunc("/data-analysis/markers", handleMarkers)
	http.HandleFunc("/data-analysis/distance-markers", handleCreateDistanceMarkers)
	http.HandleFunc("/data-analysis/trim-markers", handleTrimMarkers)
	http.HandleFunc("/data-analysis/duplicate-flight", handleDuplicateFlight)
	http.HandleFunc("/data-analysis/trim-flight", handleTrimFlight)
	http.HandleFunc("/data-analysis/delete-flight", handleDeleteFlight)
	http.HandleFunc("/data-analysis/export-csv", handleCSVExport)
	http.HandleFunc("/data-analysis/statistics", handleGetStatistics)
	http.HandleFunc("/data-analysis/api/", handleAPIRequest)
}

func serveDataAnalysisPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	DataAnalysisPage().Render(r.Context(), w)
}

func handleDatabaseUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("database")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".sdlog" && ext != ".sqlite" && ext != ".db" && ext != ".csv" {
		http.Error(w, "Invalid file format. Please upload a SQLite database file (.sdlog, .sqlite, .db) or CSV file (.csv).", http.StatusBadRequest)
		return
	}

	// Create unique filename
	timestamp := time.Now().Format("20060102_150405")
	tempFilename := fmt.Sprintf("uploaded_%s_%s", timestamp, filename)
	tempPath := filepath.Join(tempDir, tempFilename)

	// Save file
	dst, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Import flights based on file type
	var flights []Flight
	if ext == ".csv" {
		// Handle CSV import
		flight, err := importCSVFile(tempPath, filename)
		if err != nil {
			os.Remove(tempPath)
			http.Error(w, fmt.Sprintf("Failed to import CSV: %v", err), http.StatusBadRequest)
			return
		}
		flights = []Flight{*flight}
	} else {
		// Handle database import
		var err error
		flights, err = ImportFlightsFromDatabase(tempPath)
		if err != nil {
			os.Remove(tempPath)
			http.Error(w, fmt.Sprintf("Failed to import flights: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Clean up temporary file
	os.Remove(tempPath)

	response := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Successfully imported %d flights from %s", len(flights), filename),
		"flights": flights,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetFlights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flights, err := getFlightsFromMainDB()
	if err != nil {
		http.Error(w, "Failed to get flights", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flights)
}

func handleGetFlightData(w http.ResponseWriter, r *http.Request) {
	flightIdStr := r.URL.Query().Get("flightId")

	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	flightData, err := getFlightDataFromMainDB(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flight data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flightData)
}

func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Handle various API endpoints for the data analysis module
	path := strings.TrimPrefix(r.URL.Path, "/data-analysis/api/")

	switch path {
	case "health":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	case "stats":
		stats, err := getMainDatabaseStats()
		if err != nil {
			http.Error(w, "Failed to get database stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	default:
		http.Error(w, "API endpoint not found", http.StatusNotFound)
	}
}

func getFlightsFromMainDB() ([]Flight, error) {
	query := `
		SELECT id, title, flight_number, start_zulu_sim_time, end_zulu_sim_time
		FROM flight
		ORDER BY start_zulu_sim_time DESC
	`

	rows, err := mainDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var f Flight
		var title, flightNumber sql.NullString
		var startTime, endTime string

		err := rows.Scan(&f.ID, &title, &flightNumber, &startTime, &endTime)
		if err != nil {
			return nil, err
		}

		f.Title = title.String
		if f.Title == "" {
			f.Title = "Untitled"
		}

		f.FlightNumber = flightNumber.String
		if f.FlightNumber == "" {
			f.FlightNumber = "No Number"
		}

		f.StartTime = startTime
		f.EndTime = endTime

		flights = append(flights, f)
	}

	return flights, nil
}

func getFlightDataFromMainDB(flightID int) (*FlightData, error) {
	// Get flight details
	flight, err := getFlightByIDFromMainDB(flightID)
	if err != nil {
		return nil, err
	}

	// Get aircraft for this flight
	aircraft, err := getAircraftByFlightIDFromMainDB(flightID)
	if err != nil {
		return nil, err
	}

	flightData := &FlightData{
		Flight:       flight,
		PositionData: make(map[string][]PositionPoint),
		EngineData:   make(map[string][]EnginePoint),
	}

	// Get position and engine data for each aircraft
	for _, ac := range aircraft {
		// Get position data with airspeed
		positionData, err := getPositionDataWithAirspeedFromMainDB(ac.ID)
		if err != nil {
			log.Printf("Failed to get position data for aircraft %d: %v", ac.ID, err)
			continue
		}

		// Get engine data
		engineData, err := getEngineDataFromMainDB(ac.ID)
		if err != nil {
			log.Printf("Failed to get engine data for aircraft %d: %v", ac.ID, err)
		}

		aircraftLabel := ac.Type
		if ac.TailNumber != "" {
			aircraftLabel += fmt.Sprintf(" (%s)", ac.TailNumber)
		}

		if len(positionData) > 0 {
			flightData.PositionData[aircraftLabel] = positionData
		}

		if len(engineData) > 0 {
			flightData.EngineData[aircraftLabel] = engineData
		}
	}

	return flightData, nil
}

func getFlightByIDFromMainDB(flightID int) (*Flight, error) {
	query := `
		SELECT id, title, flight_number, start_zulu_sim_time, end_zulu_sim_time
		FROM flight
		WHERE id = ?
	`

	var f Flight
	var title, flightNumber sql.NullString
	var startTime, endTime string

	err := mainDB.QueryRow(query, flightID).Scan(&f.ID, &title, &flightNumber, &startTime, &endTime)
	if err != nil {
		return nil, err
	}

	f.Title = title.String
	if f.Title == "" {
		f.Title = "Untitled"
	}

	f.FlightNumber = flightNumber.String
	if f.FlightNumber == "" {
		f.FlightNumber = "No Number"
	}

	f.StartTime = startTime
	f.EndTime = endTime

	return &f, nil
}

func getAircraftByFlightIDFromMainDB(flightID int) ([]Aircraft, error) {
	query := `
		SELECT id, flight_id, seq_nr, type, tail_number, airline
		FROM aircraft
		WHERE flight_id = ?
	`

	rows, err := mainDB.Query(query, flightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aircraft []Aircraft
	for rows.Next() {
		var ac Aircraft
		var tailNumber, airline sql.NullString

		err := rows.Scan(&ac.ID, &ac.FlightID, &ac.SeqNr, &ac.Type, &tailNumber, &airline)
		if err != nil {
			return nil, err
		}

		ac.TailNumber = tailNumber.String
		ac.Airline = airline.String

		aircraft = append(aircraft, ac)
	}

	return aircraft, nil
}

func getPositionDataWithAirspeedFromMainDB(aircraftID int) ([]PositionPoint, error) {
	// Get position data
	positionQuery := `
		SELECT timestamp, altitude, latitude, longitude, 
		       indicated_altitude, pressure_altitude, indicated_airspeed
		FROM position
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	rows, err := mainDB.Query(positionQuery, aircraftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []PositionPoint
	var minTimestamp *int64

	for rows.Next() {
		var pos PositionPoint
		var timestamp int64
		var altitude, latitude, longitude sql.NullFloat64
		var indicatedAltitude, pressureAltitude, indicatedAirspeed sql.NullFloat64

		err := rows.Scan(&timestamp, &altitude, &latitude, &longitude,
			&indicatedAltitude, &pressureAltitude, &indicatedAirspeed)
		if err != nil {
			return nil, err
		}

		if minTimestamp == nil {
			minTimestamp = &timestamp
		}

		pos.Timestamp = timestamp
		pos.TimestampSeconds = float64(timestamp-*minTimestamp) / 1000.0
		pos.Altitude = altitude.Float64
		pos.Latitude = latitude.Float64
		pos.Longitude = longitude.Float64
		pos.IndicatedAltitude = indicatedAltitude.Float64
		pos.PressureAltitude = pressureAltitude.Float64
		
		// Use stored indicated airspeed when available (CSV data)
		if indicatedAirspeed.Valid && indicatedAirspeed.Float64 > 0 {
			pos.Airspeed = indicatedAirspeed.Float64
		} else {
			pos.Airspeed = 0.0 // Will be set later from attitude data if available
		}

		positions = append(positions, pos)
	}

	// Get attitude data for airspeed calculation
	attitudeQuery := `
		SELECT timestamp, velocity_x, velocity_y, velocity_z
		FROM attitude
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	attitudeRows, err := mainDB.Query(attitudeQuery, aircraftID)
	if err != nil {
		// If attitude data is not available, return positions without airspeed
		return positions, nil
	}
	defer attitudeRows.Close()

	type AttitudePoint struct {
		Timestamp        int64
		TimestampSeconds float64
		VelocityX        float64
		VelocityY        float64
		VelocityZ        float64
		Airspeed         float64
	}

	var attitudes []AttitudePoint
	for attitudeRows.Next() {
		var att AttitudePoint
		var timestamp int64
		var velocityX, velocityY, velocityZ sql.NullFloat64

		err := attitudeRows.Scan(&timestamp, &velocityX, &velocityY, &velocityZ)
		if err != nil {
			continue
		}

		if minTimestamp != nil {
			att.Timestamp = timestamp
			att.TimestampSeconds = float64(timestamp-*minTimestamp) / 1000.0
			att.VelocityX = velocityX.Float64
			att.VelocityY = velocityY.Float64
			att.VelocityZ = velocityZ.Float64

			// Calculate airspeed from velocity components
			att.Airspeed = calculateMagnitude(att.VelocityX, att.VelocityY, att.VelocityZ)

			attitudes = append(attitudes, att)
		}
	}

	// Match airspeed to position data (only for positions without stored indicated airspeed)
	for i := range positions {
		// Skip if position already has indicated airspeed from CSV data
		if positions[i].Airspeed > 0 {
			continue
		}
		
		// Find closest attitude data point for calculated airspeed
		closestAirspeed := 0.0
		minTimeDiff := float64(^uint(0) >> 1) // Max float64

		for _, att := range attitudes {
			timeDiff := abs(att.TimestampSeconds - positions[i].TimestampSeconds)
			if timeDiff < minTimeDiff {
				minTimeDiff = timeDiff
				closestAirspeed = att.Airspeed
			}
		}

		positions[i].Airspeed = closestAirspeed
	}

	return positions, nil
}

func getEngineDataFromMainDB(aircraftID int) ([]EnginePoint, error) {
	query := `
		SELECT timestamp, 
		       throttle_lever_position1, throttle_lever_position2, 
		       throttle_lever_position3, throttle_lever_position4
		FROM engine
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	rows, err := mainDB.Query(query, aircraftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engines []EnginePoint
	var minTimestamp *int64

	for rows.Next() {
		var eng EnginePoint
		var timestamp int64
		var throttle1, throttle2, throttle3, throttle4 sql.NullFloat64

		err := rows.Scan(&timestamp, &throttle1, &throttle2, &throttle3, &throttle4)
		if err != nil {
			return nil, err
		}

		if minTimestamp == nil {
			minTimestamp = &timestamp
		}

		eng.Timestamp = timestamp
		eng.TimestampSeconds = float64(timestamp-*minTimestamp) / 1000.0
		eng.ThrottlePosition1 = throttle1.Float64
		eng.ThrottlePosition2 = throttle2.Float64
		eng.ThrottlePosition3 = throttle3.Float64
		eng.ThrottlePosition4 = throttle4.Float64

		engines = append(engines, eng)
	}

	return engines, nil
}

func getMainDatabaseStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get flight count
	var flightCount int
	err := mainDB.QueryRow("SELECT COUNT(*) FROM flight").Scan(&flightCount)
	if err != nil {
		return nil, err
	}
	stats["flight_count"] = flightCount

	// Get aircraft count
	var aircraftCount int
	err = mainDB.QueryRow("SELECT COUNT(*) FROM aircraft").Scan(&aircraftCount)
	if err != nil {
		return nil, err
	}
	stats["aircraft_count"] = aircraftCount

	// Get position data points count
	var positionCount int
	err = mainDB.QueryRow("SELECT COUNT(*) FROM position").Scan(&positionCount)
	if err != nil {
		return nil, err
	}
	stats["position_count"] = positionCount

	// Get database file size
	if fileInfo, err := os.Stat(mainDatabasePath); err == nil {
		stats["database_size_bytes"] = fileInfo.Size()
		stats["database_size_mb"] = float64(fileInfo.Size()) / (1024 * 1024)
	}

	return stats, nil
}

// Helper functions
func calculateMagnitude(x, y, z float64) float64 {
	return sqrt(x*x + y*y + z*z)
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Marker database functions
func getMarkersForFlight(flightID int) ([]Marker, error) {
	query := `
		SELECT id, flight_id, time_seconds, label, COALESCE(type, 'regular'), created_at
		FROM markers
		WHERE flight_id = ?
		ORDER BY time_seconds
	`

	rows, err := mainDB.Query(query, flightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markers []Marker
	for rows.Next() {
		var m Marker
		err := rows.Scan(&m.ID, &m.FlightID, &m.Time, &m.Label, &m.Type, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		markers = append(markers, m)
	}

	return markers, nil
}

func createMarker(marker Marker) (*Marker, error) {
	// Set default type if not specified
	if marker.Type == "" {
		marker.Type = "regular"
	}

	query := `
		INSERT INTO markers (flight_id, time_seconds, label, type)
		VALUES (?, ?, ?, ?)
	`

	result, err := mainDB.Exec(query, marker.FlightID, marker.Time, marker.Label, marker.Type)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Return the created marker with ID
	createdMarker := &Marker{
		ID:       int(id),
		FlightID: marker.FlightID,
		Time:     marker.Time,
		Label:    marker.Label,
		Type:     marker.Type,
	}

	return createdMarker, nil
}

func deleteMarker(markerID int) error {
	query := `DELETE FROM markers WHERE id = ?`
	_, err := mainDB.Exec(query, markerID)
	return err
}

// Marker HTTP handlers
func handleMarkers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetMarkers(w, r)
	case http.MethodPost:
		handleCreateMarker(w, r)
	case http.MethodDelete:
		handleDeleteMarker(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetMarkers(w http.ResponseWriter, r *http.Request) {
	flightIdStr := r.URL.Query().Get("flightId")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	markers, err := getMarkersForFlight(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get markers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(markers)
}

func handleCreateMarker(w http.ResponseWriter, r *http.Request) {
	var marker Marker
	if err := json.NewDecoder(r.Body).Decode(&marker); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if marker.FlightID == 0 || marker.Label == "" {
		http.Error(w, "Flight ID and label are required", http.StatusBadRequest)
		return
	}

	createdMarker, err := createMarker(marker)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create marker: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdMarker)
}

func handleDeleteMarker(w http.ResponseWriter, r *http.Request) {
	markerIdStr := r.URL.Query().Get("id")
	if markerIdStr == "" {
		http.Error(w, "Marker ID required", http.StatusBadRequest)
		return
	}

	markerId, err := strconv.Atoi(markerIdStr)
	if err != nil {
		http.Error(w, "Invalid marker ID", http.StatusBadRequest)
		return
	}

	if err := deleteMarker(markerId); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete marker: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Trim marker specialized functions

// createOrUpdateTrimMarker creates or updates a trim marker (trim_start or trim_end)
func createOrUpdateTrimMarker(flightID int, markerType string, time float64, label string) (*Marker, error) {
	if markerType != "trim_start" && markerType != "trim_end" {
		return nil, fmt.Errorf("invalid trim marker type: %s", markerType)
	}

	// Check if trim marker of this type already exists
	existingMarker, err := getTrimMarker(flightID, markerType)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if existingMarker != nil {
		// Update existing marker
		query := `UPDATE markers SET time_seconds = ?, label = ? WHERE id = ?`
		_, err := mainDB.Exec(query, time, label, existingMarker.ID)
		if err != nil {
			return nil, err
		}
		existingMarker.Time = time
		existingMarker.Label = label
		return existingMarker, nil
	} else {
		// Create new marker
		marker := Marker{
			FlightID: flightID,
			Time:     time,
			Label:    label,
			Type:     markerType,
		}
		return createMarker(marker)
	}
}

// getTrimMarker retrieves a specific trim marker (trim_start or trim_end) for a flight
func getTrimMarker(flightID int, markerType string) (*Marker, error) {
	query := `
		SELECT id, flight_id, time_seconds, label, type, created_at
		FROM markers
		WHERE flight_id = ? AND type = ?
	`

	var m Marker
	err := mainDB.QueryRow(query, flightID, markerType).Scan(&m.ID, &m.FlightID, &m.Time, &m.Label, &m.Type, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// getTrimMarkers retrieves both trim markers for a flight
func getTrimMarkers(flightID int) (trimStart *Marker, trimEnd *Marker, err error) {
	trimStart, _ = getTrimMarker(flightID, "trim_start")
	trimEnd, _ = getTrimMarker(flightID, "trim_end")
	return trimStart, trimEnd, nil
}

// deleteTrimMarkers removes all trim markers for a flight
func deleteTrimMarkers(flightID int) error {
	query := `DELETE FROM markers WHERE flight_id = ? AND type IN ('trim_start', 'trim_end')`
	_, err := mainDB.Exec(query, flightID)
	return err
}

// HTTP handler for trim markers
func handleTrimMarkers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetTrimMarkers(w, r)
	case http.MethodPost:
		handleCreateTrimMarker(w, r)
	case http.MethodDelete:
		handleDeleteTrimMarkers(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetTrimMarkers(w http.ResponseWriter, r *http.Request) {
	flightIdStr := r.URL.Query().Get("flightId")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	trimStart, trimEnd, err := getTrimMarkers(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get trim markers: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]*Marker{
		"trim_start": trimStart,
		"trim_end":   trimEnd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateTrimMarker(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FlightID int     `json:"flight_id"`
		Type     string  `json:"type"`      // "trim_start" or "trim_end"
		Time     float64 `json:"time"`
		Label    string  `json:"label"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.FlightID == 0 || request.Type == "" {
		http.Error(w, "Flight ID and type are required", http.StatusBadRequest)
		return
	}

	marker, err := createOrUpdateTrimMarker(request.FlightID, request.Type, request.Time, request.Label)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create trim marker: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marker)
}

func handleDeleteTrimMarkers(w http.ResponseWriter, r *http.Request) {
	flightIdStr := r.URL.Query().Get("flightId")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	if err := deleteTrimMarkers(flightId); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete trim markers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// calculateDistanceNM calculates the distance between two points in nautical miles
func calculateDistanceNM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 3440.065 // Earth's radius in nautical miles
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// findDistanceMarkers analyzes position data to find the first point where aircraft reaches exactly targetDistanceNM from Currock Hill
func findDistanceMarkers(positionData []PositionPoint) []float64 {
	var markerTimes []float64
	var prevDistance float64
	var prevTime float64
	markerFound := false
	const tolerance = 0.05 // 0.05 nm tolerance for "exactly" 9nm

	for i, pos := range positionData {
		if pos.Latitude == 0 && pos.Longitude == 0 {
			continue // Skip invalid coordinates
		}

		distance := calculateDistanceNM(pos.Latitude, pos.Longitude, currockHillLat, currockHillLon)
		
		if i > 0 && !markerFound {
			// Check if we crossed the target distance (from either direction)
			if (prevDistance > targetDistanceNM && distance <= targetDistanceNM) ||
			   (prevDistance < targetDistanceNM && distance >= targetDistanceNM) {
				// Interpolate the exact crossing time
				if prevDistance != distance {
					ratio := (targetDistanceNM - prevDistance) / (distance - prevDistance)
					crossingTime := prevTime + ratio*(pos.TimestampSeconds-prevTime)
					markerTimes = append(markerTimes, crossingTime)
					markerFound = true // Only create one marker per aircraft
				}
			}
			// Also check if we're very close to exactly the target distance
			if math.Abs(distance-targetDistanceNM) <= tolerance {
				markerTimes = append(markerTimes, pos.TimestampSeconds)
				markerFound = true
			}
		}
		
		prevDistance = distance
		prevTime = pos.TimestampSeconds
	}

	return markerTimes
}

// createDistanceMarkersForFlight automatically creates distance markers for a flight
func createDistanceMarkersForFlight(flightID int) error {
	// Get flight data
	flightData, err := getFlightDataFromMainDB(flightID)
	if err != nil {
		return fmt.Errorf("failed to get flight data: %v", err)
	}

	// Process each aircraft's position data
	for aircraftLabel, positionData := range flightData.PositionData {
		markerTimes := findDistanceMarkers(positionData)
		
		for _, markerTime := range markerTimes {
			label := fmt.Sprintf("9nm from Currock Hill - %s", aircraftLabel)
			
			marker := Marker{
				FlightID: flightID,
				Time:     markerTime,
				Label:    label,
			}
			
			_, err := createMarker(marker)
			if err != nil {
				log.Printf("Failed to create distance marker: %v", err)
				continue
			}
			
			log.Printf("Created distance marker at %.2fs for flight %d: %s", markerTime, flightID, label)
		}
	}

	return nil
}

// HTTP handler for creating distance markers
func handleCreateDistanceMarkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flightIdStr := r.URL.Query().Get("flightId")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	err = createDistanceMarkersForFlight(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create distance markers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Distance markers created successfully",
	})
}

// HTTP handler for duplicating flights
func handleDuplicateFlight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var request struct {
		FlightID int    `json:"flight_id"`
		NewTitle string `json:"new_title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if request.FlightID == 0 || request.NewTitle == "" {
		http.Error(w, "Flight ID and new title are required", http.StatusBadRequest)
		return
	}

	// Check if title already exists
	exists, err := flightTitleExists(request.NewTitle)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check title uniqueness: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "A flight with this title already exists", http.StatusConflict)
		return
	}

	// Duplicate the flight
	newFlightID, err := duplicateFlight(request.FlightID, request.NewTitle)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to duplicate flight: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"message":       fmt.Sprintf("Flight duplicated successfully with ID %d", newFlightID),
		"new_flight_id": newFlightID,
	})
}

// flightTitleExists checks if a flight title already exists in the database
func flightTitleExists(title string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM flight WHERE title = ?"
	err := mainDB.QueryRow(query, title).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// duplicateFlight duplicates a flight with all its related data
func duplicateFlight(originalFlightID int, newTitle string) (int, error) {
	// Start transaction
	tx, err := mainDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Copy the flight record
	newFlightID, err := duplicateFlightRecord(tx, originalFlightID, newTitle)
	if err != nil {
		return 0, fmt.Errorf("failed to duplicate flight record: %w", err)
	}

	// Step 2: Get all aircraft for the original flight
	aircraft, err := getAircraftByFlightIDFromMainDB(originalFlightID)
	if err != nil {
		return 0, fmt.Errorf("failed to get aircraft: %w", err)
	}

	// Step 3: Duplicate each aircraft and its data
	for _, ac := range aircraft {
		newAircraftID, err := duplicateAircraftRecord(tx, ac, newFlightID)
		if err != nil {
			return 0, fmt.Errorf("failed to duplicate aircraft %d: %w", ac.ID, err)
		}

		// Duplicate all related data for this aircraft
		if err := duplicatePositionData(tx, ac.ID, newAircraftID); err != nil {
			return 0, fmt.Errorf("failed to duplicate position data for aircraft %d: %w", ac.ID, err)
		}

		if err := duplicateAttitudeData(tx, ac.ID, newAircraftID); err != nil {
			return 0, fmt.Errorf("failed to duplicate attitude data for aircraft %d: %w", ac.ID, err)
		}

		if err := duplicateEngineData(tx, ac.ID, newAircraftID); err != nil {
			return 0, fmt.Errorf("failed to duplicate engine data for aircraft %d: %w", ac.ID, err)
		}
	}

	// Step 4: Duplicate markers
	if err := duplicateMarkers(tx, originalFlightID, newFlightID); err != nil {
		return 0, fmt.Errorf("failed to duplicate markers: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully duplicated flight %d as flight %d with title '%s'", originalFlightID, newFlightID, newTitle)
	return newFlightID, nil
}

// duplicateFlightRecord copies a flight record with a new title
func duplicateFlightRecord(tx *sql.Tx, originalFlightID int, newTitle string) (int, error) {
	// Get original flight data
	query := `
		SELECT title, flight_number, start_zulu_sim_time, end_zulu_sim_time,
		       description, user_aircraft_seq_nr, surface_type, surface_condition,
		       on_any_runway, on_parking_spot, ground_altitude, ambient_temperature,
		       total_air_temperature, wind_speed, wind_direction, visibility,
		       sea_level_pressure, pitot_icing, structural_icing, precipitation_state,
		       in_clouds, start_local_sim_time, end_local_sim_time
		FROM flight WHERE id = ?
	`
	
	var originalTitle, flightNumber, description sql.NullString
	var startZulu, endZulu, startLocal, endLocal string
	var userAircraftSeqNr, surfaceType, surfaceCondition sql.NullInt64
	var onAnyRunway, onParkingSpot, inClouds sql.NullInt64
	var groundAltitude, ambientTemp, totalAirTemp, windSpeed, windDirection sql.NullFloat64
	var visibility, seaLevelPressure, pitotIcing, structuralIcing sql.NullFloat64
	var precipitationState sql.NullInt64

	err := tx.QueryRow(query, originalFlightID).Scan(
		&originalTitle, &flightNumber, &startZulu, &endZulu,
		&description, &userAircraftSeqNr, &surfaceType, &surfaceCondition,
		&onAnyRunway, &onParkingSpot, &groundAltitude, &ambientTemp,
		&totalAirTemp, &windSpeed, &windDirection, &visibility,
		&seaLevelPressure, &pitotIcing, &structuralIcing, &precipitationState,
		&inClouds, &startLocal, &endLocal,
	)
	if err != nil {
		return 0, err
	}

	// Insert new flight record with new title
	insertQuery := `
		INSERT INTO flight (
			title, flight_number, start_zulu_sim_time, end_zulu_sim_time,
			description, user_aircraft_seq_nr, surface_type, surface_condition,
			on_any_runway, on_parking_spot, ground_altitude, ambient_temperature,
			total_air_temperature, wind_speed, wind_direction, visibility,
			sea_level_pressure, pitot_icing, structural_icing, precipitation_state,
			in_clouds, start_local_sim_time, end_local_sim_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(insertQuery,
		newTitle, flightNumber, startZulu, endZulu,
		description, userAircraftSeqNr, surfaceType, surfaceCondition,
		onAnyRunway, onParkingSpot, groundAltitude, ambientTemp,
		totalAirTemp, windSpeed, windDirection, visibility,
		seaLevelPressure, pitotIcing, structuralIcing, precipitationState,
		inClouds, startLocal, endLocal,
	)
	if err != nil {
		return 0, err
	}

	newFlightID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(newFlightID), nil
}

// duplicateAircraftRecord copies an aircraft record for a new flight
func duplicateAircraftRecord(tx *sql.Tx, aircraft Aircraft, newFlightID int) (int, error) {
	query := `
		SELECT seq_nr, type, time_offset, tail_number, airline,
		       initial_airspeed, altitude_above_ground, start_on_ground
		FROM aircraft WHERE id = ?
	`
	
	var seqNr, timeOffset sql.NullInt64
	var aircraftType string
	var tailNumber, airline sql.NullString
	var initialAirspeed sql.NullInt64
	var altitudeAboveGround sql.NullFloat64
	var startOnGround sql.NullInt64

	err := tx.QueryRow(query, aircraft.ID).Scan(
		&seqNr, &aircraftType, &timeOffset, &tailNumber, &airline,
		&initialAirspeed, &altitudeAboveGround, &startOnGround,
	)
	if err != nil {
		return 0, err
	}

	insertQuery := `
		INSERT INTO aircraft (
			flight_id, seq_nr, type, time_offset, tail_number, airline,
			initial_airspeed, altitude_above_ground, start_on_ground
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(insertQuery,
		newFlightID, seqNr, aircraftType, timeOffset, tailNumber, airline,
		initialAirspeed, altitudeAboveGround, startOnGround,
	)
	if err != nil {
		return 0, err
	}

	newAircraftID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(newAircraftID), nil
}

// duplicatePositionData copies all position data for an aircraft
func duplicatePositionData(tx *sql.Tx, originalAircraftID, newAircraftID int) error {
	query := `
		SELECT timestamp, latitude, longitude, altitude, indicated_altitude,
		       calibrated_indicated_altitude, pressure_altitude, indicated_airspeed
		FROM position WHERE aircraft_id = ? ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO position (
			aircraft_id, timestamp, latitude, longitude, altitude,
			indicated_altitude, calibrated_indicated_altitude, pressure_altitude, indicated_airspeed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var latitude, longitude, altitude sql.NullFloat64
		var indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude, indicatedAirspeed sql.NullFloat64

		err := rows.Scan(
			&timestamp, &latitude, &longitude, &altitude,
			&indicatedAltitude, &calibratedIndicatedAltitude, &pressureAltitude, &indicatedAirspeed,
		)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			newAircraftID, timestamp, latitude, longitude, altitude,
			indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude, indicatedAirspeed,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateAttitudeData copies all attitude data for an aircraft
func duplicateAttitudeData(tx *sql.Tx, originalAircraftID, newAircraftID int) error {
	query := `
		SELECT timestamp, pitch, bank, true_heading, velocity_x, velocity_y, velocity_z, on_ground
		FROM attitude WHERE aircraft_id = ? ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO attitude (
			aircraft_id, timestamp, pitch, bank, true_heading,
			velocity_x, velocity_y, velocity_z, on_ground
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var pitch, bank, trueHeading sql.NullFloat64
		var velocityX, velocityY, velocityZ sql.NullFloat64
		var onGround sql.NullInt64

		err := rows.Scan(
			&timestamp, &pitch, &bank, &trueHeading,
			&velocityX, &velocityY, &velocityZ, &onGround,
		)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			newAircraftID, timestamp, pitch, bank, trueHeading,
			velocityX, velocityY, velocityZ, onGround,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateEngineData copies all engine data for an aircraft
func duplicateEngineData(tx *sql.Tx, originalAircraftID, newAircraftID int) error {
	query := `
		SELECT timestamp, throttle_lever_position1, throttle_lever_position2,
		       throttle_lever_position3, throttle_lever_position4,
		       propeller_lever_position1, propeller_lever_position2,
		       propeller_lever_position3, propeller_lever_position4,
		       mixture_lever_position1, mixture_lever_position2,
		       mixture_lever_position3, mixture_lever_position4,
		       cowl_flap_position1, cowl_flap_position2,
		       cowl_flap_position3, cowl_flap_position4,
		       electrical_master_battery1, electrical_master_battery2,
		       electrical_master_battery3, electrical_master_battery4,
		       general_engine_starter1, general_engine_starter2,
		       general_engine_starter3, general_engine_starter4,
		       general_engine_combustion1, general_engine_combustion2,
		       general_engine_combustion3, general_engine_combustion4
		FROM engine WHERE aircraft_id = ? ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO engine (
			aircraft_id, timestamp, throttle_lever_position1, throttle_lever_position2,
			throttle_lever_position3, throttle_lever_position4,
			propeller_lever_position1, propeller_lever_position2,
			propeller_lever_position3, propeller_lever_position4,
			mixture_lever_position1, mixture_lever_position2,
			mixture_lever_position3, mixture_lever_position4,
			cowl_flap_position1, cowl_flap_position2,
			cowl_flap_position3, cowl_flap_position4,
			electrical_master_battery1, electrical_master_battery2,
			electrical_master_battery3, electrical_master_battery4,
			general_engine_starter1, general_engine_starter2,
			general_engine_starter3, general_engine_starter4,
			general_engine_combustion1, general_engine_combustion2,
			general_engine_combustion3, general_engine_combustion4
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var throttle1, throttle2, throttle3, throttle4 sql.NullFloat64
		var prop1, prop2, prop3, prop4 sql.NullFloat64
		var mixture1, mixture2, mixture3, mixture4 sql.NullFloat64
		var cowl1, cowl2, cowl3, cowl4 sql.NullFloat64
		var battery1, battery2, battery3, battery4 sql.NullInt64
		var starter1, starter2, starter3, starter4 sql.NullInt64
		var combustion1, combustion2, combustion3, combustion4 sql.NullInt64

		err := rows.Scan(
			&timestamp, &throttle1, &throttle2, &throttle3, &throttle4,
			&prop1, &prop2, &prop3, &prop4,
			&mixture1, &mixture2, &mixture3, &mixture4,
			&cowl1, &cowl2, &cowl3, &cowl4,
			&battery1, &battery2, &battery3, &battery4,
			&starter1, &starter2, &starter3, &starter4,
			&combustion1, &combustion2, &combustion3, &combustion4,
		)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			newAircraftID, timestamp, throttle1, throttle2, throttle3, throttle4,
			prop1, prop2, prop3, prop4,
			mixture1, mixture2, mixture3, mixture4,
			cowl1, cowl2, cowl3, cowl4,
			battery1, battery2, battery3, battery4,
			starter1, starter2, starter3, starter4,
			combustion1, combustion2, combustion3, combustion4,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateMarkers copies all markers for a flight
func duplicateMarkers(tx *sql.Tx, originalFlightID, newFlightID int) error {
	query := `
		SELECT time_seconds, label
		FROM markers WHERE flight_id = ?
		ORDER BY time_seconds
	`

	rows, err := tx.Query(query, originalFlightID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO markers (flight_id, time_seconds, label)
		VALUES (?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timeSeconds float64
		var label string

		err := rows.Scan(&timeSeconds, &label)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(newFlightID, timeSeconds, label)
		if err != nil {
			return err
		}
	}

	return nil
}

// HTTP handler for trimming flights
func handleTrimFlight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var request struct {
		FlightID  int     `json:"flight_id"`
		NewTitle  string  `json:"new_title"`
		StartTime float64 `json:"start_time"`
		EndTime   float64 `json:"end_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if request.FlightID == 0 || request.NewTitle == "" {
		http.Error(w, "Flight ID and new title are required", http.StatusBadRequest)
		return
	}

	if request.EndTime <= request.StartTime {
		http.Error(w, "End time must be greater than start time", http.StatusBadRequest)
		return
	}

	if request.EndTime-request.StartTime < 1.0 {
		http.Error(w, "Trim range too small (minimum 1 second)", http.StatusBadRequest)
		return
	}

	// Check if title already exists
	exists, err := flightTitleExists(request.NewTitle)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check title uniqueness: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "A flight with this title already exists", http.StatusConflict)
		return
	}

	// Trim the flight
	newFlightID, err := trimFlight(request.FlightID, request.NewTitle, request.StartTime, request.EndTime)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to trim flight: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"message":       fmt.Sprintf("Flight trimmed successfully with ID %d", newFlightID),
		"new_flight_id": newFlightID,
	})
}

// trimFlight trims a flight to a specific time range
func trimFlight(originalFlightID int, newTitle string, startTime, endTime float64) (int, error) {
	// Start transaction
	tx, err := mainDB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Copy the flight record
	newFlightID, err := duplicateFlightRecord(tx, originalFlightID, newTitle)
	if err != nil {
		return 0, fmt.Errorf("failed to duplicate flight record: %w", err)
	}

	// Step 2: Get all aircraft for the original flight
	aircraft, err := getAircraftByFlightIDFromMainDB(originalFlightID)
	if err != nil {
		return 0, fmt.Errorf("failed to get aircraft: %w", err)
	}

	// Step 3: Duplicate each aircraft and its trimmed data
	for _, ac := range aircraft {
		newAircraftID, err := duplicateAircraftRecord(tx, ac, newFlightID)
		if err != nil {
			return 0, fmt.Errorf("failed to duplicate aircraft %d: %w", ac.ID, err)
		}

		// Duplicate all related data for this aircraft with time filtering
		if err := duplicatePositionDataTrimmed(tx, ac.ID, newAircraftID, startTime, endTime); err != nil {
			return 0, fmt.Errorf("failed to duplicate position data for aircraft %d: %w", ac.ID, err)
		}

		if err := duplicateAttitudeDataTrimmed(tx, ac.ID, newAircraftID, startTime, endTime); err != nil {
			return 0, fmt.Errorf("failed to duplicate attitude data for aircraft %d: %w", ac.ID, err)
		}

		if err := duplicateEngineDataTrimmed(tx, ac.ID, newAircraftID, startTime, endTime); err != nil {
			return 0, fmt.Errorf("failed to duplicate engine data for aircraft %d: %w", ac.ID, err)
		}
	}

	// Step 4: Duplicate markers within the trim range
	if err := duplicateMarkersTrimmed(tx, originalFlightID, newFlightID, startTime, endTime); err != nil {
		return 0, fmt.Errorf("failed to duplicate markers: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully trimmed flight %d to time range %.1f-%.1fs as flight %d with title '%s'", originalFlightID, startTime, endTime, newFlightID, newTitle)
	return newFlightID, nil
}

// duplicatePositionDataTrimmed copies position data within a specific time range, adjusting timestamps to start from 0
func duplicatePositionDataTrimmed(tx *sql.Tx, originalAircraftID, newAircraftID int, startTime, endTime float64) error {
	// Calculate the minimum timestamp to normalize timestamps to start from 0
	var minTimestamp int64
	err := tx.QueryRow("SELECT MIN(timestamp) FROM position WHERE aircraft_id = ?", originalAircraftID).Scan(&minTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Convert time range to milliseconds and add to base timestamp
	startTimestamp := minTimestamp + int64(startTime*1000)
	endTimestamp := minTimestamp + int64(endTime*1000)

	query := `
		SELECT timestamp, latitude, longitude, altitude, indicated_altitude,
		       calibrated_indicated_altitude, pressure_altitude, indicated_airspeed
		FROM position 
		WHERE aircraft_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID, startTimestamp, endTimestamp)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO position (
			aircraft_id, timestamp, latitude, longitude, altitude,
			indicated_altitude, calibrated_indicated_altitude, pressure_altitude, indicated_airspeed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var latitude, longitude, altitude sql.NullFloat64
		var indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude, indicatedAirspeed sql.NullFloat64

		err := rows.Scan(
			&timestamp, &latitude, &longitude, &altitude,
			&indicatedAltitude, &calibratedIndicatedAltitude, &pressureAltitude, &indicatedAirspeed,
		)
		if err != nil {
			return err
		}

		// Adjust timestamp to start from the new base (startTimestamp becomes minTimestamp)
		adjustedTimestamp := minTimestamp + (timestamp - startTimestamp)

		_, err = stmt.Exec(
			newAircraftID, adjustedTimestamp, latitude, longitude, altitude,
			indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude, indicatedAirspeed,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateAttitudeDataTrimmed copies attitude data within a specific time range, adjusting timestamps to start from 0
func duplicateAttitudeDataTrimmed(tx *sql.Tx, originalAircraftID, newAircraftID int, startTime, endTime float64) error {
	// Calculate the minimum timestamp to normalize timestamps to start from 0
	var minTimestamp int64
	err := tx.QueryRow("SELECT MIN(timestamp) FROM attitude WHERE aircraft_id = ?", originalAircraftID).Scan(&minTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Convert time range to milliseconds and add to base timestamp
	startTimestamp := minTimestamp + int64(startTime*1000)
	endTimestamp := minTimestamp + int64(endTime*1000)

	query := `
		SELECT timestamp, pitch, bank, true_heading, velocity_x, velocity_y, velocity_z, on_ground
		FROM attitude 
		WHERE aircraft_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID, startTimestamp, endTimestamp)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO attitude (
			aircraft_id, timestamp, pitch, bank, true_heading,
			velocity_x, velocity_y, velocity_z, on_ground
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var pitch, bank, trueHeading sql.NullFloat64
		var velocityX, velocityY, velocityZ sql.NullFloat64
		var onGround sql.NullInt64

		err := rows.Scan(
			&timestamp, &pitch, &bank, &trueHeading,
			&velocityX, &velocityY, &velocityZ, &onGround,
		)
		if err != nil {
			return err
		}

		// Adjust timestamp to start from the new base
		adjustedTimestamp := minTimestamp + (timestamp - startTimestamp)

		_, err = stmt.Exec(
			newAircraftID, adjustedTimestamp, pitch, bank, trueHeading,
			velocityX, velocityY, velocityZ, onGround,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateEngineDataTrimmed copies engine data within a specific time range, adjusting timestamps to start from 0
func duplicateEngineDataTrimmed(tx *sql.Tx, originalAircraftID, newAircraftID int, startTime, endTime float64) error {
	// Calculate the minimum timestamp to normalize timestamps to start from 0
	var minTimestamp int64
	err := tx.QueryRow("SELECT MIN(timestamp) FROM engine WHERE aircraft_id = ?", originalAircraftID).Scan(&minTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Convert time range to milliseconds and add to base timestamp
	startTimestamp := minTimestamp + int64(startTime*1000)
	endTimestamp := minTimestamp + int64(endTime*1000)

	query := `
		SELECT timestamp, throttle_lever_position1, throttle_lever_position2,
		       throttle_lever_position3, throttle_lever_position4,
		       propeller_lever_position1, propeller_lever_position2,
		       propeller_lever_position3, propeller_lever_position4,
		       mixture_lever_position1, mixture_lever_position2,
		       mixture_lever_position3, mixture_lever_position4,
		       cowl_flap_position1, cowl_flap_position2,
		       cowl_flap_position3, cowl_flap_position4,
		       electrical_master_battery1, electrical_master_battery2,
		       electrical_master_battery3, electrical_master_battery4,
		       general_engine_starter1, general_engine_starter2,
		       general_engine_starter3, general_engine_starter4,
		       general_engine_combustion1, general_engine_combustion2,
		       general_engine_combustion3, general_engine_combustion4
		FROM engine 
		WHERE aircraft_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp
	`

	rows, err := tx.Query(query, originalAircraftID, startTimestamp, endTimestamp)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO engine (
			aircraft_id, timestamp, throttle_lever_position1, throttle_lever_position2,
			throttle_lever_position3, throttle_lever_position4,
			propeller_lever_position1, propeller_lever_position2,
			propeller_lever_position3, propeller_lever_position4,
			mixture_lever_position1, mixture_lever_position2,
			mixture_lever_position3, mixture_lever_position4,
			cowl_flap_position1, cowl_flap_position2,
			cowl_flap_position3, cowl_flap_position4,
			electrical_master_battery1, electrical_master_battery2,
			electrical_master_battery3, electrical_master_battery4,
			general_engine_starter1, general_engine_starter2,
			general_engine_starter3, general_engine_starter4,
			general_engine_combustion1, general_engine_combustion2,
			general_engine_combustion3, general_engine_combustion4
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var throttle1, throttle2, throttle3, throttle4 sql.NullFloat64
		var prop1, prop2, prop3, prop4 sql.NullFloat64
		var mixture1, mixture2, mixture3, mixture4 sql.NullFloat64
		var cowl1, cowl2, cowl3, cowl4 sql.NullFloat64
		var battery1, battery2, battery3, battery4 sql.NullInt64
		var starter1, starter2, starter3, starter4 sql.NullInt64
		var combustion1, combustion2, combustion3, combustion4 sql.NullInt64

		err := rows.Scan(
			&timestamp, &throttle1, &throttle2, &throttle3, &throttle4,
			&prop1, &prop2, &prop3, &prop4,
			&mixture1, &mixture2, &mixture3, &mixture4,
			&cowl1, &cowl2, &cowl3, &cowl4,
			&battery1, &battery2, &battery3, &battery4,
			&starter1, &starter2, &starter3, &starter4,
			&combustion1, &combustion2, &combustion3, &combustion4,
		)
		if err != nil {
			return err
		}

		// Adjust timestamp to start from the new base
		adjustedTimestamp := minTimestamp + (timestamp - startTimestamp)

		_, err = stmt.Exec(
			newAircraftID, adjustedTimestamp, throttle1, throttle2, throttle3, throttle4,
			prop1, prop2, prop3, prop4,
			mixture1, mixture2, mixture3, mixture4,
			cowl1, cowl2, cowl3, cowl4,
			battery1, battery2, battery3, battery4,
			starter1, starter2, starter3, starter4,
			combustion1, combustion2, combustion3, combustion4,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// duplicateMarkersTrimmed copies markers within a specific time range, adjusting time to start from 0
func duplicateMarkersTrimmed(tx *sql.Tx, originalFlightID, newFlightID int, startTime, endTime float64) error {
	query := `
		SELECT time_seconds, label, COALESCE(type, 'regular')
		FROM markers 
		WHERE flight_id = ? AND time_seconds >= ? AND time_seconds <= ?
		ORDER BY time_seconds
	`

	rows, err := tx.Query(query, originalFlightID, startTime, endTime)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO markers (flight_id, time_seconds, label, type)
		VALUES (?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timeSeconds float64
		var label, markerType string

		err := rows.Scan(&timeSeconds, &label, &markerType)
		if err != nil {
			return err
		}

		// Adjust time to start from 0
		adjustedTime := timeSeconds - startTime

		_, err = stmt.Exec(newFlightID, adjustedTime, label, markerType)
		if err != nil {
			return err
		}
	}

	return nil
}

// handleGetStatistics handles requests for flight data statistics
func handleGetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flightIdStr := r.URL.Query().Get("flightId")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	// Get flight data
	flightData, err := getFlightDataFromMainDB(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flight data: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	statistics := CalculateFlightStatistics(flightData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statistics)
}

// importCSVFile imports flight data from a CSV file
func importCSVFile(filePath, filename string) (*Flight, error) {
	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Validate CSV structure first
	if err := ValidateCSVStructure(file); err != nil {
		return nil, fmt.Errorf("invalid CSV structure: %w", err)
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Parse CSV with default options
	options := CSVImportOptions{
		FlightTitle:  extractFlightTitle(filename),
		AircraftType: "Unknown",
		SkipRows:     2, // Skip separator and comment rows
	}

	csvData, err := ParseCSVFlightData(file, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	// Import into database
	flight, err := ImportFlightFromCSV(csvData)
	if err != nil {
		return nil, fmt.Errorf("failed to import CSV to database: %w", err)
	}

	return flight, nil
}

// extractFlightTitle extracts a meaningful flight title from filename
func extractFlightTitle(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Remove common prefixes/suffixes
	name = strings.TrimPrefix(name, "uploaded_")
	name = strings.ReplaceAll(name, "_", " ")
	
	// Capitalize first letter
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	
	if name == "" {
		name = "CSV Flight Data"
	}
	
	return name
}

// handleDeleteFlight handles flight deletion requests
func handleDeleteFlight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flightIdStr := r.URL.Query().Get("id")
	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	// Get flight title for logging before deletion
	flight, err := getFlightByIDFromMainDB(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Flight not found: %v", err), http.StatusNotFound)
		return
	}

	// Delete the flight
	if err := DeleteFlight(flightId); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete flight: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Flight '%s' (ID: %d) deleted successfully", flight.Title, flightId),
	})
}
