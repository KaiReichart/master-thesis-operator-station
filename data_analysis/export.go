package data_analysis

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// CSVExportOptions defines options for CSV export
type CSVExportOptions struct {
	FlightID int
	Format   string // "airspeed-altitude", "full"
}

// ExportFlightDataToCSV exports flight data to ZIP file containing two CSV files
func ExportFlightDataToCSV(flightData *FlightData, options CSVExportOptions) (*bytes.Buffer, error) {
	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)

	// Create a new zip archive
	w := zip.NewWriter(buf)

	// Generate airspeed CSV
	airspeedData, err := generateAirspeedCSV(flightData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate airspeed CSV: %w", err)
	}

	// Generate altitude CSV
	altitudeData, err := generateAltitudeCSV(flightData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate altitude CSV: %w", err)
	}

	// Add airspeed CSV to zip
	airspeedFile, err := w.Create("airspeed_data.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create airspeed CSV file in zip: %w", err)
	}
	if _, err := airspeedFile.Write(airspeedData); err != nil {
		return nil, fmt.Errorf("failed to write airspeed CSV data: %w", err)
	}

	// Add altitude CSV to zip
	altitudeFile, err := w.Create("altitude_data.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create altitude CSV file in zip: %w", err)
	}
	if _, err := altitudeFile.Write(altitudeData); err != nil {
		return nil, fmt.Errorf("failed to write altitude CSV data: %w", err)
	}

	// Close the zip writer
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf, nil
}

// generateAirspeedCSV generates CSV data for airspeed information (IAS only)
func generateAirspeedCSV(flightData *FlightData) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Write header
	header := []string{"Timestamp", "IAS"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows - combine all aircraft data
	for _, positionData := range flightData.PositionData {
		for _, point := range positionData {
			row := []string{
				fmt.Sprintf("%.1f", point.TimestampSeconds),
				fmt.Sprintf("%.2f", point.Airspeed),
			}
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// generateAltitudeCSV generates CSV data for altitude information (essential data only)
func generateAltitudeCSV(flightData *FlightData) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// Write header
	header := []string{"Timestamp", "Altitude"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows - combine all aircraft data, use MSL altitude as primary
	for _, positionData := range flightData.PositionData {
		for _, point := range positionData {
			row := []string{
				fmt.Sprintf("%.1f", point.TimestampSeconds),
				fmt.Sprintf("%.2f", point.Altitude),
			}
			if err := writer.Write(row); err != nil {
				return nil, fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateCSVFilename generates a filename for the CSV export ZIP
func GenerateCSVFilename(flight *Flight, format string) string {
	timestamp := time.Now().Format("20060102_150405")
	flightTitle := flight.Title
	if flightTitle == "" {
		flightTitle = "Flight_" + strconv.Itoa(flight.ID)
	}

	formatSuffix := ""
	if format == "airspeed-altitude" {
		formatSuffix = "_airspeed_altitude"
	} else if format == "full" {
		formatSuffix = "_full_data"
	}

	return fmt.Sprintf("%s%s_%s.zip", flightTitle, formatSuffix, timestamp)
}

// handleCSVExport handles HTTP requests for CSV export
func handleCSVExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get parameters
	flightIdStr := r.URL.Query().Get("flightId")
	format := r.URL.Query().Get("format")

	if flightIdStr == "" {
		http.Error(w, "Flight ID required", http.StatusBadRequest)
		return
	}

	flightId, err := strconv.Atoi(flightIdStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	// Default format if not specified
	if format == "" {
		format = "airspeed-altitude"
	}

	// Validate format
	if format != "airspeed-altitude" && format != "full" {
		http.Error(w, "Invalid format. Use 'airspeed-altitude' or 'full'", http.StatusBadRequest)
		return
	}

	// Get flight data
	flightData, err := getFlightDataFromMainDB(flightId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get flight data: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate CSV ZIP file
	options := CSVExportOptions{
		FlightID: flightId,
		Format:   format,
	}

	csvBuffer, err := ExportFlightDataToCSV(flightData, options)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate CSV files: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate filename
	filename := GenerateCSVFilename(flightData.Flight, format)

	// Set headers for file download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(csvBuffer.Len()))

	// Write the ZIP file to response
	_, err = w.Write(csvBuffer.Bytes())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write CSV file: %v", err), http.StatusInternalServerError)
		return
	}
}
