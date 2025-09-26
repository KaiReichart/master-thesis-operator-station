package data_analysis

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ParseCSVFlightData parses a CSV file and returns structured flight data
func ParseCSVFlightData(reader io.Reader, options CSVImportOptions) (*CSVFlightData, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields
	
	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) < 3 {
		return nil, fmt.Errorf("CSV file too short, expected at least 3 rows (metadata, header, data)")
	}
	
	// Parse metadata from the first few rows
	metadata, err := parseCSVMetadata(records, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	
	// Find header row (contains column names)
	headerRowIndex := -1
	var headers []string
	for i, record := range records {
		if i < options.SkipRows {
			continue
		}
		// Look for row that contains "Time" and other expected columns
		if containsFlightDataHeaders(record) {
			headerRowIndex = i
			headers = record
			break
		}
	}
	
	if headerRowIndex == -1 {
		return nil, fmt.Errorf("could not find header row with flight data columns")
	}
	
	// Parse data records
	var flightRecords []CSVFlightRecord
	startTime := time.Time{}
	
	for i := headerRowIndex + 1; i < len(records); i++ {
		record := records[i]
		if len(record) != len(headers) {
			continue // Skip malformed rows
		}
		
		flightRecord, err := parseCSVRecord(headers, record)
		if err != nil {
			// Log error but continue with other records
			continue
		}
		
		// Calculate relative timestamp in seconds
		if recordTime, err := time.Parse("2006-01-02T15:04:05.9999999-07:00", flightRecord.Time); err == nil {
			if startTime.IsZero() {
				startTime = recordTime
				flightRecord.TimestampSeconds = 0
			} else {
				flightRecord.TimestampSeconds = recordTime.Sub(startTime).Seconds()
			}
		}
		
		flightRecords = append(flightRecords, *flightRecord)
	}
	
	if len(flightRecords) == 0 {
		return nil, fmt.Errorf("no valid flight data records found")
	}
	
	metadata.TotalRecords = len(flightRecords)
	
	return &CSVFlightData{
		Metadata: *metadata,
		Headers:  headers,
		Records:  flightRecords,
	}, nil
}

// parseCSVMetadata extracts metadata from the first few rows of the CSV
func parseCSVMetadata(records [][]string, options CSVImportOptions) (*CSVMetadata, error) {
	metadata := &CSVMetadata{
		FlightTitle:  options.FlightTitle,
		AircraftType: options.AircraftType,
	}
	
	// Look for metadata in first few rows
	for i := 0; i < len(records) && i < 5; i++ {
		if len(records[i]) == 0 {
			continue
		}
		
		row := strings.Join(records[i], " ")
		
		// Extract source information
		if strings.Contains(row, "FS-FlightControl") {
			metadata.Source = "FS-FlightControl"
		}
		
		// Extract recording timestamp
		if strings.Contains(row, "Recorded at:") {
			// Extract timestamp from format: "Recorded at: 7/30/2025 9:05:41 PM"
			parts := strings.Split(row, "Recorded at:")
			if len(parts) > 1 {
				timeStr := strings.TrimSpace(parts[1])
				timeStr = strings.Split(timeStr, " (more info")[0] // Remove trailing info
				metadata.RecordedAt = timeStr
			}
		}
	}
	
	// Set default title if not provided
	if metadata.FlightTitle == "" {
		if metadata.RecordedAt != "" {
			metadata.FlightTitle = fmt.Sprintf("Flight %s", metadata.RecordedAt)
		} else {
			metadata.FlightTitle = "Imported CSV Flight"
		}
	}
	
	// Set default aircraft type if not provided
	if metadata.AircraftType == "" {
		metadata.AircraftType = "Unknown"
	}
	
	return metadata, nil
}

// containsFlightDataHeaders checks if a row contains expected flight data headers
func containsFlightDataHeaders(record []string) bool {
	expectedHeaders := []string{"Time", "Altitude", "Latitude", "Longitude"}
	foundCount := 0
	
	for _, header := range record {
		headerLower := strings.ToLower(header)
		for _, expected := range expectedHeaders {
			if strings.Contains(headerLower, strings.ToLower(expected)) {
				foundCount++
				break
			}
		}
	}
	
	return foundCount >= 3 // At least 3 expected headers should be present
}

// parseCSVRecord parses a single CSV record into a CSVFlightRecord
func parseCSVRecord(headers []string, record []string) (*CSVFlightRecord, error) {
	if len(headers) != len(record) {
		return nil, fmt.Errorf("header/record length mismatch")
	}
	
	flightRecord := &CSVFlightRecord{}
	
	for i, header := range headers {
		value := strings.TrimSpace(record[i])
		if value == "" {
			continue
		}
		
		headerLower := strings.ToLower(header)
		
		switch {
		case strings.Contains(header, "Time"):
			flightRecord.Time = value
			
		case strings.Contains(headerLower, "airspeedindicated"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.AirspeedIndicated = val
			}
			
		case strings.Contains(headerLower, "airspeedtrue"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.AirspeedTrue = val
			}
			
		case strings.Contains(headerLower, "groundspeed"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.GroundSpeed = val
			}
			
		case strings.Contains(headerLower, "altitude") && strings.Contains(headerLower, "feet"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.Altitude = val
			}
			
		case strings.Contains(headerLower, "groundelevation"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.GroundElevation = val
			}
			
		case strings.Contains(headerLower, "latitude"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.Latitude = val
			}
			
		case strings.Contains(headerLower, "longitude"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.Longitude = val
			}
			
		case strings.Contains(headerLower, "bankangle"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.BankAngle = val
			}
			
		case strings.Contains(headerLower, "pitchangle"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.PitchAngle = val
			}
			
		case strings.Contains(headerLower, "headingmagnetic"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.HeadingMagnetic = val
			}
			
		case strings.Contains(headerLower, "headingtrue"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.HeadingTrue = val
			}
			
		case strings.Contains(headerLower, "ambienttemperature") && !strings.Contains(headerLower, "total"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.AmbientTemperature = val
			}
			
		case strings.Contains(headerLower, "ambientwinddirection"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.AmbientWindDirection = val
			}
			
		case strings.Contains(headerLower, "ambientwindvelocity"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.AmbientWindVelocity = val
			}
			
		case strings.Contains(headerLower, "flapshandleposition"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.FlapsHandlePosition = val
			}
			
		case strings.Contains(headerLower, "fueltotalquantity"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.FuelTotalQuantity = val
			}
			
		case strings.Contains(headerLower, "geardown"):
			flightRecord.GearDown = parseBool(value)
			
		case strings.Contains(headerLower, "onground"):
			flightRecord.OnGround = parseBool(value)
			
		case strings.Contains(headerLower, "gforce"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.GForce = val
			}
			
		case strings.Contains(headerLower, "verticalspeed"):
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				flightRecord.VerticalSpeed = val
			}
			
		case strings.Contains(headerLower, "overspeedwarning"):
			flightRecord.OverspeedWarning = parseBool(value)
			
		case strings.Contains(headerLower, "stallwarning"):
			flightRecord.StallWarning = parseBool(value)
		}
	}
	
	return flightRecord, nil
}

// parseBool parses boolean values from CSV (handles "True"/"False" strings)
func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes"
}

// ValidateCSVStructure validates that the CSV has the required structure for flight data
func ValidateCSVStructure(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	
	// Read first few records to validate structure
	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}
	
	if len(records) < 3 {
		return fmt.Errorf("CSV file too short, expected at least 3 rows")
	}
	
	// Check for header row
	headerFound := false
	for _, record := range records {
		if containsFlightDataHeaders(record) {
			headerFound = true
			break
		}
	}
	
	if !headerFound {
		return fmt.Errorf("no valid flight data headers found in CSV")
	}
	
	return nil
}