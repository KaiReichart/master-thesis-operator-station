package gps

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseXGPSPacket(data []byte) (GPSData, error) {
	var gps GPSData

	// Convert the data to a string and split by commas
	parts := strings.Split(string(data), ",")
	if len(parts) < 6 {
		return gps, fmt.Errorf("invalid data format: expected at least 6 parts, got %d", len(parts))
	}

	// Parse the values
	lon, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return gps, fmt.Errorf("error parsing longitude: %v", err)
	}
	lat, err := strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return gps, fmt.Errorf("error parsing latitude: %v", err)
	}
	alt, err := strconv.ParseFloat(parts[3], 32)
	if err != nil {
		return gps, fmt.Errorf("error parsing altitude: %v", err)
	}
	hdg, err := strconv.ParseFloat(parts[4], 32)
	if err != nil {
		return gps, fmt.Errorf("error parsing heading: %v", err)
	}
	spd, err := strconv.ParseFloat(parts[5], 32)
	if err != nil {
		return gps, fmt.Errorf("error parsing speed: %v", err)
	}

	gps.Longitude = float32(lon)
	gps.Latitude = float32(lat)
	gps.AltitudeMSL = float32(alt)
	gps.TrueHeading = float32(hdg)
	gps.GroundSpeed = float32(spd)

	return gps, nil
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
