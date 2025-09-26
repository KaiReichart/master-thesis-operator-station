package data_analysis

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	mainDatabasePath = "data/data_analysis.db"
)

var (
	mainDB *sql.DB
)

// InitMainDatabase initializes the main data analysis database
func InitMainDatabase() error {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	var err error
	mainDB, err = sql.Open("sqlite3", mainDatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open main database: %w", err)
	}

	// Test connection
	if err := mainDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping main database: %w", err)
	}

	// Create schema if it doesn't exist
	if err := createMainDatabaseSchema(); err != nil {
		return fmt.Errorf("failed to create main database schema: %w", err)
	}

	log.Println("Main data analysis database initialized successfully")
	return nil
}

// createMainDatabaseSchema creates the necessary tables in the main database
func createMainDatabaseSchema() error {
	// Check if the database is already initialized by looking for a key table
	var count int
	err := mainDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='flight'").Scan(&count)
	if err == nil && count > 0 {
		// Database already initialized, but check if markers table exists
		log.Println("Main database schema already exists, checking for markers table...")
		if err := ensureMarkersTable(); err != nil {
			return err
		}
		return ensurePositionTableColumns()
	}

	log.Println("Initializing main database schema...")

	// Read the schema from structure.sql
	schemaPath := filepath.Join("data", "structure.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute the schema
	_, err = mainDB.Exec(string(schemaBytes))
	if err != nil {
		// If there's an error, it might be because tables already exist
		// Let's check if the essential tables exist
		var flightCount, aircraftCount, positionCount int
		mainDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='flight'").Scan(&flightCount)
		mainDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='aircraft'").Scan(&aircraftCount)
		mainDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='position'").Scan(&positionCount)

		if flightCount > 0 && aircraftCount > 0 && positionCount > 0 {
			// Essential tables exist, schema is probably fine
			log.Println("Essential database tables already exist, continuing...")
			// Still need to ensure markers table exists
			if err := ensureMarkersTable(); err != nil {
				return err
			}
			return ensurePositionTableColumns()
		}

		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Main database schema created successfully")
	// Create markers table
	if err := ensureMarkersTable(); err != nil {
		return err
	}
	return ensurePositionTableColumns()
}

// ensureMarkersTable creates the markers table if it doesn't exist
func ensureMarkersTable() error {
	var count int
	err := mainDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='markers'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check markers table: %w", err)
	}
	
	if count > 0 {
		log.Println("Markers table already exists, checking for type column...")
		return ensureMarkerTypeColumn()
	}
	
	log.Println("Creating markers table...")
	markersSchema := `
		CREATE TABLE markers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			flight_id INTEGER NOT NULL,
			time_seconds REAL NOT NULL,
			label TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'regular',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(flight_id) REFERENCES flight(id) ON DELETE CASCADE
		);
		
		CREATE INDEX markers_flight_id_idx ON markers (flight_id);
		CREATE INDEX markers_time_idx ON markers (flight_id, time_seconds);
		CREATE INDEX markers_type_idx ON markers (flight_id, type);
	`
	
	_, err = mainDB.Exec(markersSchema)
	if err != nil {
		return fmt.Errorf("failed to create markers table: %w", err)
	}
	
	log.Println("Markers table created successfully")
	return nil
}

// ensureMarkerTypeColumn adds the type column to existing markers table if it doesn't exist
func ensureMarkerTypeColumn() error {
	// Check if type column exists
	var typeColumnExists bool
	rows, err := mainDB.Query("PRAGMA table_info(markers)")
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue sql.NullString
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			return fmt.Errorf("failed to scan table info: %w", err)
		}
		
		if name == "type" {
			typeColumnExists = true
			break
		}
	}

	if typeColumnExists {
		log.Println("Marker type column already exists")
		return nil
	}

	log.Println("Adding type column to markers table...")
	
	// Add the type column with default value
	_, err = mainDB.Exec("ALTER TABLE markers ADD COLUMN type TEXT NOT NULL DEFAULT 'regular'")
	if err != nil {
		return fmt.Errorf("failed to add type column: %w", err)
	}

	// Create index for the new column
	_, err = mainDB.Exec("CREATE INDEX IF NOT EXISTS markers_type_idx ON markers (flight_id, type)")
	if err != nil {
		return fmt.Errorf("failed to create type index: %w", err)
	}

	log.Println("Marker type column added successfully")
	return nil
}

// ensurePositionTableColumns ensures the position table has all required columns
func ensurePositionTableColumns() error {
	// Check if indicated_airspeed column exists
	var indicatedAirspeedExists bool
	rows, err := mainDB.Query("PRAGMA table_info(position)")
	if err != nil {
		return fmt.Errorf("failed to get position table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dfltValue sql.NullString
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			return fmt.Errorf("failed to scan position table info: %w", err)
		}
		
		if name == "indicated_airspeed" {
			indicatedAirspeedExists = true
			break
		}
	}

	if indicatedAirspeedExists {
		log.Println("Position table indicated_airspeed column already exists")
		return nil
	}

	log.Println("Adding indicated_airspeed column to position table...")
	
	// Add the indicated_airspeed column
	_, err = mainDB.Exec("ALTER TABLE position ADD COLUMN indicated_airspeed REAL")
	if err != nil {
		return fmt.Errorf("failed to add indicated_airspeed column: %w", err)
	}

	log.Println("Position table indicated_airspeed column added successfully")
	return nil
}

// GetMainDatabase returns the main database connection
func GetMainDatabase() *sql.DB {
	return mainDB
}

// CloseMainDatabase closes the main database connection
func CloseMainDatabase() error {
	if mainDB != nil {
		return mainDB.Close()
	}
	return nil
}

// ImportFlightsFromDatabase imports all flights and related data from an uploaded database
func ImportFlightsFromDatabase(sourceDBPath string) ([]Flight, error) {
	// Open the source database
	sourceDB, err := sql.Open("sqlite3", sourceDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceDB.Close()

	// Verify source database has required tables
	if err := verifyDatabaseSchema(sourceDB); err != nil {
		return nil, fmt.Errorf("invalid source database: %w", err)
	}

	// Start transaction
	tx, err := mainDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Import flights
	flights, err := importFlights(sourceDB, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to import flights: %w", err)
	}

	// Import aircraft for each flight
	for _, flight := range flights {
		if err := importAircraftForFlight(sourceDB, tx, flight.SourceID, flight.ID); err != nil {
			return nil, fmt.Errorf("failed to import aircraft for flight %d: %w", flight.SourceID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully imported %d flights from %s", len(flights), sourceDBPath)
	return flights, nil
}

// verifyDatabaseSchema verifies that the source database has the required schema
func verifyDatabaseSchema(db *sql.DB) error {
	requiredTables := []string{"flight", "aircraft", "position", "attitude", "engine"}

	for _, table := range requiredTables {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&tableName)
		if err != nil {
			return fmt.Errorf("required table '%s' not found", table)
		}
	}

	return nil
}

// importFlights imports flight records from source database to main database
func importFlights(sourceDB *sql.DB, tx *sql.Tx) ([]Flight, error) {
	query := `
		SELECT id, title, flight_number, start_zulu_sim_time, end_zulu_sim_time,
		       description, user_aircraft_seq_nr, surface_type, surface_condition,
		       on_any_runway, on_parking_spot, ground_altitude, ambient_temperature,
		       total_air_temperature, wind_speed, wind_direction, visibility,
		       sea_level_pressure, pitot_icing, structural_icing, precipitation_state,
		       in_clouds, start_local_sim_time, end_local_sim_time
		FROM flight
		ORDER BY start_zulu_sim_time DESC
	`

	rows, err := sourceDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	var flights []Flight
	for rows.Next() {
		var sourceID int
		var title, flightNumber, description sql.NullString
		var startZulu, endZulu, startLocal, endLocal string
		var userAircraftSeqNr, surfaceType, surfaceCondition sql.NullInt64
		var onAnyRunway, onParkingSpot, inClouds sql.NullInt64
		var groundAltitude, ambientTemp, totalAirTemp, windSpeed, windDirection sql.NullFloat64
		var visibility, seaLevelPressure, pitotIcing, structuralIcing sql.NullFloat64
		var precipitationState sql.NullInt64

		err := rows.Scan(
			&sourceID, &title, &flightNumber, &startZulu, &endZulu,
			&description, &userAircraftSeqNr, &surfaceType, &surfaceCondition,
			&onAnyRunway, &onParkingSpot, &groundAltitude, &ambientTemp,
			&totalAirTemp, &windSpeed, &windDirection, &visibility,
			&seaLevelPressure, &pitotIcing, &structuralIcing, &precipitationState,
			&inClouds, &startLocal, &endLocal,
		)
		if err != nil {
			return nil, err
		}

		result, err := tx.Exec(insertQuery,
			title, flightNumber, startZulu, endZulu,
			description, userAircraftSeqNr, surfaceType, surfaceCondition,
			onAnyRunway, onParkingSpot, groundAltitude, ambientTemp,
			totalAirTemp, windSpeed, windDirection, visibility,
			seaLevelPressure, pitotIcing, structuralIcing, precipitationState,
			inClouds, startLocal, endLocal,
		)
		if err != nil {
			return nil, err
		}

		newID, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		flight := Flight{
			ID:           int(newID),
			SourceID:     sourceID,
			Title:        title.String,
			FlightNumber: flightNumber.String,
			StartTime:    startZulu,
			EndTime:      endZulu,
		}

		if flight.Title == "" {
			flight.Title = "Untitled"
		}
		if flight.FlightNumber == "" {
			flight.FlightNumber = "No Number"
		}

		flights = append(flights, flight)
	}

	return flights, nil
}

// importAircraftForFlight imports aircraft and all related data for a specific flight
func importAircraftForFlight(sourceDB *sql.DB, tx *sql.Tx, sourceFlightID, newFlightID int) error {
	// Get aircraft for this flight
	aircraftQuery := `
		SELECT id, seq_nr, type, time_offset, tail_number, airline,
		       initial_airspeed, altitude_above_ground, start_on_ground
		FROM aircraft
		WHERE flight_id = ?
	`

	rows, err := sourceDB.Query(aircraftQuery, sourceFlightID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertAircraftQuery := `
		INSERT INTO aircraft (
			flight_id, seq_nr, type, time_offset, tail_number, airline,
			initial_airspeed, altitude_above_ground, start_on_ground
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	for rows.Next() {
		var sourceAircraftID, seqNr sql.NullInt64
		var aircraftType string
		var timeOffset sql.NullInt64
		var tailNumber, airline sql.NullString
		var initialAirspeed sql.NullInt64
		var altitudeAboveGround sql.NullFloat64
		var startOnGround sql.NullInt64

		err := rows.Scan(
			&sourceAircraftID, &seqNr, &aircraftType, &timeOffset,
			&tailNumber, &airline, &initialAirspeed, &altitudeAboveGround, &startOnGround,
		)
		if err != nil {
			return err
		}

		result, err := tx.Exec(insertAircraftQuery,
			newFlightID, seqNr, aircraftType, timeOffset,
			tailNumber, airline, initialAirspeed, altitudeAboveGround, startOnGround,
		)
		if err != nil {
			return err
		}

		newAircraftID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Import position data
		if err := importPositionData(sourceDB, tx, int(sourceAircraftID.Int64), int(newAircraftID)); err != nil {
			return fmt.Errorf("failed to import position data: %w", err)
		}

		// Import attitude data
		if err := importAttitudeData(sourceDB, tx, int(sourceAircraftID.Int64), int(newAircraftID)); err != nil {
			return fmt.Errorf("failed to import attitude data: %w", err)
		}

		// Import engine data
		if err := importEngineData(sourceDB, tx, int(sourceAircraftID.Int64), int(newAircraftID)); err != nil {
			return fmt.Errorf("failed to import engine data: %w", err)
		}
	}

	return nil
}

// importPositionData imports position data for an aircraft
func importPositionData(sourceDB *sql.DB, tx *sql.Tx, sourceAircraftID, newAircraftID int) error {
	query := `
		SELECT timestamp, latitude, longitude, altitude, indicated_altitude,
		       calibrated_indicated_altitude, pressure_altitude
		FROM position
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	rows, err := sourceDB.Query(query, sourceAircraftID)
	if err != nil {
		return err
	}
	defer rows.Close()

	insertQuery := `
		INSERT INTO position (
			aircraft_id, timestamp, latitude, longitude, altitude,
			indicated_altitude, calibrated_indicated_altitude, pressure_altitude
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var timestamp int64
		var latitude, longitude, altitude sql.NullFloat64
		var indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude sql.NullFloat64

		err := rows.Scan(
			&timestamp, &latitude, &longitude, &altitude,
			&indicatedAltitude, &calibratedIndicatedAltitude, &pressureAltitude,
		)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			newAircraftID, timestamp, latitude, longitude, altitude,
			indicatedAltitude, calibratedIndicatedAltitude, pressureAltitude,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// importAttitudeData imports attitude data for an aircraft
func importAttitudeData(sourceDB *sql.DB, tx *sql.Tx, sourceAircraftID, newAircraftID int) error {
	query := `
		SELECT timestamp, pitch, bank, true_heading, velocity_x, velocity_y, velocity_z, on_ground
		FROM attitude
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	rows, err := sourceDB.Query(query, sourceAircraftID)
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

// importEngineData imports engine data for an aircraft
func importEngineData(sourceDB *sql.DB, tx *sql.Tx, sourceAircraftID, newAircraftID int) error {
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
		WHERE aircraft_id = ?
		ORDER BY timestamp
	`

	rows, err := sourceDB.Query(query, sourceAircraftID)
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

// ImportFlightFromCSV imports flight data from parsed CSV data
func ImportFlightFromCSV(csvData *CSVFlightData) (*Flight, error) {
	if len(csvData.Records) == 0 {
		return nil, fmt.Errorf("no flight data records to import")
	}

	// Start transaction
	tx, err := mainDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create flight record
	flightID, err := createFlightFromCSV(tx, csvData)
	if err != nil {
		return nil, fmt.Errorf("failed to create flight: %w", err)
	}

	// Create aircraft record
	aircraftID, err := createAircraftFromCSV(tx, flightID, csvData)
	if err != nil {
		return nil, fmt.Errorf("failed to create aircraft: %w", err)
	}

	// Import position data
	if err := importPositionDataFromCSV(tx, aircraftID, csvData); err != nil {
		return nil, fmt.Errorf("failed to import position data: %w", err)
	}

	// Import attitude data
	if err := importAttitudeDataFromCSV(tx, aircraftID, csvData); err != nil {
		return nil, fmt.Errorf("failed to import attitude data: %w", err)
	}

	// Import engine data (limited data available from CSV)
	if err := importEngineDataFromCSV(tx, aircraftID, csvData); err != nil {
		return nil, fmt.Errorf("failed to import engine data: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created flight
	flight := &Flight{
		ID:          flightID,
		Title:       csvData.Metadata.FlightTitle,
		FlightNumber: "CSV Import",
		StartTime:   csvData.Metadata.RecordedAt,
		EndTime:     csvData.Metadata.RecordedAt,
	}

	log.Printf("Successfully imported CSV flight: %s (%d records)", flight.Title, len(csvData.Records))
	return flight, nil
}

// createFlightFromCSV creates a flight record from CSV metadata
func createFlightFromCSV(tx *sql.Tx, csvData *CSVFlightData) (int, error) {
	// Create flight times from first and last records
	var startTime, endTime string
	if len(csvData.Records) > 0 {
		startTime = csvData.Records[0].Time
		endTime = csvData.Records[len(csvData.Records)-1].Time
	}

	query := `
		INSERT INTO flight (
			title, flight_number, start_zulu_sim_time, end_zulu_sim_time,
			description, user_aircraft_seq_nr
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	description := fmt.Sprintf("Imported from CSV (%s) - %d data points", 
		csvData.Metadata.Source, csvData.Metadata.TotalRecords)

	result, err := tx.Exec(query,
		csvData.Metadata.FlightTitle,
		"CSV Import",
		startTime,
		endTime,
		description,
		1, // user_aircraft_seq_nr - default to 1 for CSV data
	)
	if err != nil {
		return 0, err
	}

	flightID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(flightID), nil
}

// createAircraftFromCSV creates an aircraft record from CSV data
func createAircraftFromCSV(tx *sql.Tx, flightID int, csvData *CSVFlightData) (int, error) {
	query := `
		INSERT INTO aircraft (
			flight_id, seq_nr, type, tail_number
		) VALUES (?, ?, ?, ?)
	`

	result, err := tx.Exec(query,
		flightID,
		1, // Single aircraft for CSV data
		csvData.Metadata.AircraftType,
		"CSV-IMPORT",
	)
	if err != nil {
		return 0, err
	}

	aircraftID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(aircraftID), nil
}

// importPositionDataFromCSV imports position data from CSV records
func importPositionDataFromCSV(tx *sql.Tx, aircraftID int, csvData *CSVFlightData) error {
	query := `
		INSERT INTO position (
			aircraft_id, timestamp, latitude, longitude, altitude,
			indicated_altitude, pressure_altitude, indicated_airspeed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Calculate base timestamp from first record
	var baseTimestamp int64
	if len(csvData.Records) > 0 {
		// Use milliseconds since epoch, with relative timing
		baseTimestamp = 1690000000000 // Arbitrary base timestamp
	}

	for _, record := range csvData.Records {
		// Convert timestamp to milliseconds
		timestamp := baseTimestamp + int64(record.TimestampSeconds*1000)
		
		// Convert altitude from feet to meters for consistency
		altitudeMeters := record.Altitude * 0.3048
		
		_, err = stmt.Exec(
			aircraftID,
			timestamp,
			record.Latitude,
			record.Longitude,
			altitudeMeters,
			record.Altitude, // Keep indicated altitude in feet
			record.Altitude, // Use same for pressure altitude
			record.AirspeedIndicated, // Store indicated airspeed in knots
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// importAttitudeDataFromCSV imports attitude data from CSV records
func importAttitudeDataFromCSV(tx *sql.Tx, aircraftID int, csvData *CSVFlightData) error {
	query := `
		INSERT INTO attitude (
			aircraft_id, timestamp, pitch, bank, true_heading,
			velocity_x, velocity_y, velocity_z, on_ground
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Calculate base timestamp from first record
	var baseTimestamp int64
	if len(csvData.Records) > 0 {
		baseTimestamp = 1690000000000 // Arbitrary base timestamp
	}

	for _, record := range csvData.Records {
		timestamp := baseTimestamp + int64(record.TimestampSeconds*1000)
		
		// Calculate velocity components from ground speed and heading
		groundSpeedMS := record.GroundSpeed * 0.514444 // knots to m/s
		headingRad := record.HeadingTrue * 3.14159 / 180.0
		
		velocityX := groundSpeedMS * sin(headingRad)
		velocityY := groundSpeedMS * cos(headingRad)
		velocityZ := record.VerticalSpeed * 0.00508 // ft/min to m/s
		
		onGround := 0
		if record.OnGround {
			onGround = 1
		}

		_, err = stmt.Exec(
			aircraftID,
			timestamp,
			record.PitchAngle,
			record.BankAngle,
			record.HeadingTrue,
			velocityX,
			velocityY,
			velocityZ,
			onGround,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// importEngineDataFromCSV imports limited engine data from CSV records
func importEngineDataFromCSV(tx *sql.Tx, aircraftID int, csvData *CSVFlightData) error {
	query := `
		INSERT INTO engine (
			aircraft_id, timestamp, throttle_lever_position1
		) VALUES (?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Calculate base timestamp from first record
	var baseTimestamp int64
	if len(csvData.Records) > 0 {
		baseTimestamp = 1690000000000 // Arbitrary base timestamp
	}

	for _, record := range csvData.Records {
		timestamp := baseTimestamp + int64(record.TimestampSeconds*1000)
		
		// Use flaps position as a proxy for throttle data (limited CSV data)
		throttlePosition := record.FlapsHandlePosition / 100.0 // Normalize to 0-1
		
		_, err = stmt.Exec(
			aircraftID,
			timestamp,
			throttlePosition,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Simple sin function implementation for velocity calculations
func sin(x float64) float64 {
	// Simple approximation using Taylor series
	x = x - 2*3.14159*float64(int(x/(2*3.14159))) // Normalize to [-2π, 2π]
	return x - (x*x*x)/6 + (x*x*x*x*x)/120
}

// Simple cos function implementation for velocity calculations  
func cos(x float64) float64 {
	return sin(x + 3.14159/2)
}

// DeleteFlight deletes a flight and all associated data
func DeleteFlight(flightID int) error {
	if flightID <= 0 {
		return fmt.Errorf("invalid flight ID: %d", flightID)
	}

	// Start transaction
	tx, err := mainDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, get all aircraft IDs for this flight
	aircraftIDs, err := getAircraftIDsForFlight(tx, flightID)
	if err != nil {
		return fmt.Errorf("failed to get aircraft IDs: %w", err)
	}

	// Delete data for each aircraft (in reverse order due to foreign keys)
	for _, aircraftID := range aircraftIDs {
		// Delete position data
		if _, err := tx.Exec("DELETE FROM position WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete position data for aircraft %d: %w", aircraftID, err)
		}

		// Delete attitude data
		if _, err := tx.Exec("DELETE FROM attitude WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete attitude data for aircraft %d: %w", aircraftID, err)
		}

		// Delete engine data
		if _, err := tx.Exec("DELETE FROM engine WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete engine data for aircraft %d: %w", aircraftID, err)
		}

		// Delete other aircraft-related data
		if _, err := tx.Exec("DELETE FROM handle WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete handle data for aircraft %d: %w", aircraftID, err)
		}

		if _, err := tx.Exec("DELETE FROM light WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete light data for aircraft %d: %w", aircraftID, err)
		}

		if _, err := tx.Exec("DELETE FROM primary_flight_control WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete primary flight control data for aircraft %d: %w", aircraftID, err)
		}

		if _, err := tx.Exec("DELETE FROM secondary_flight_control WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete secondary flight control data for aircraft %d: %w", aircraftID, err)
		}

		if _, err := tx.Exec("DELETE FROM waypoint WHERE aircraft_id = ?", aircraftID); err != nil {
			return fmt.Errorf("failed to delete waypoint data for aircraft %d: %w", aircraftID, err)
		}
	}

	// Delete markers for this flight
	if _, err := tx.Exec("DELETE FROM markers WHERE flight_id = ?", flightID); err != nil {
		return fmt.Errorf("failed to delete markers for flight %d: %w", flightID, err)
	}

	// Delete aircraft records
	if _, err := tx.Exec("DELETE FROM aircraft WHERE flight_id = ?", flightID); err != nil {
		return fmt.Errorf("failed to delete aircraft for flight %d: %w", flightID, err)
	}

	// Finally, delete the flight record
	result, err := tx.Exec("DELETE FROM flight WHERE id = ?", flightID)
	if err != nil {
		return fmt.Errorf("failed to delete flight %d: %w", flightID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flight with ID %d not found", flightID)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	log.Printf("Successfully deleted flight %d with all associated data", flightID)
	return nil
}

// getAircraftIDsForFlight retrieves all aircraft IDs associated with a flight
func getAircraftIDsForFlight(tx *sql.Tx, flightID int) ([]int, error) {
	rows, err := tx.Query("SELECT id FROM aircraft WHERE flight_id = ?", flightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aircraftIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		aircraftIDs = append(aircraftIDs, id)
	}

	return aircraftIDs, nil
}