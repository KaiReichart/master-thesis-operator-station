package gps

import "time"

// Position represents GPS position data
type Position struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude"`
	Timestamp time.Time `json:"timestamp"`
}

// Config represents GPS configuration
type Config struct {
	TargetIP          string  `json:"target_ip"`
	DistanceThreshold float64 `json:"distance_threshold"`
	IsSending         bool    `json:"is_sending"`
}

// GPSData represents the position information from an XGPS packet
type GPSData struct {
	Latitude      float32
	Longitude     float32
	AltitudeMSL   float32
	GroundSpeed   float32
	TrueHeading   float32
	MagHeading    float32
	IAS           float32
	TAS           float32
	VerticalSpeed float32
}
