package gps

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/kaireichart/master-thesis-operator-station/events"
)

//go:generate go tool templ generate

// HTML templates

func SetupHandlers() {
	http.HandleFunc("/gps/position", handleGPSPosition)
	http.HandleFunc("/gps/config", handleGPSConfig)
	http.HandleFunc("/gps/set-target-ip", handleSetTargetIPHTMX)
	http.HandleFunc("/gps/set-distance-threshold", handleSetDistanceThresholdHTMX)
	http.HandleFunc("/gps/broadcast-toggle", handleBroadcastToggleHTMX)
}

// HTMX Handlers

func handleGPSPosition(w http.ResponseWriter, r *http.Request) {
	position := GetCurrentPosition()

	w.Header().Set("Content-Type", "text/html")
	err := GPSPosition(position).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGPSConfig(w http.ResponseWriter, r *http.Request) {
	ip := GetTargetIP()
	threshold := GetDistanceThreshold()
	sending := IsSendingToTarget()

	config := &Config{
		TargetIP:          ip,
		DistanceThreshold: threshold,
		IsSending:         sending,
	}

	w.Header().Set("Content-Type", "text/html")
	err := GPSConfig(config).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSetTargetIPHTMX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.FormValue("target_ip")
	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	// Validate IP address
	if net.ParseIP(ip) == nil {
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	targetIPMutex.Lock()
	targetIP = ip
	targetIPMutex.Unlock()

	// Create and record the event
	event := events.Event{
		Type:      "target_ip_set",
		Program:   "GPS",
		Timestamp: time.Now(),
	}
	events.LogEvent(event)

	// Return updated config
	handleGPSConfig(w, r)
}

func handleSetDistanceThresholdHTMX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	thresholdStr := r.FormValue("distance_threshold")
	if thresholdStr == "" {
		http.Error(w, "Distance threshold is required", http.StatusBadRequest)
		return
	}

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil || threshold <= 0 {
		http.Error(w, "Invalid distance threshold", http.StatusBadRequest)
		return
	}

	maxDistanceMux.Lock()
	maxDistanceNM = threshold
	maxDistanceMux.Unlock()

	// Create and record the event
	event := events.Event{
		Type:      "distance_threshold_updated",
		Program:   "GPS",
		Timestamp: time.Now(),
	}
	events.LogEvent(event)

	// Return updated config
	handleGPSConfig(w, r)
}

func handleBroadcastToggleHTMX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sendingMutex.Lock()
	isSendingToTarget = !isSendingToTarget
	newState := isSendingToTarget
	sendingMutex.Unlock()

	// Create and record the event
	event := events.Event{
		Type:      "sending_toggled",
		Program:   "GPS",
		Timestamp: time.Now(),
	}
	events.LogEvent(event)

	w.Header().Set("Content-Type", "text/html")
	err := BroadcastToggle(newState).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper functions for templates

func degreesToDMS(decimalDegrees float64, isLatitude bool) string {
	absolute := math.Abs(decimalDegrees)

	degrees := int(absolute)
	minutesNotTruncated := (absolute - float64(degrees)) * 60
	minutes := int(minutesNotTruncated)
	seconds := (minutesNotTruncated - float64(minutes)) * 60

	var direction string
	if isLatitude {
		if decimalDegrees >= 0 {
			direction = "N"
		} else {
			direction = "S"
		}
	} else {
		if decimalDegrees >= 0 {
			direction = "E"
		} else {
			direction = "W"
		}
	}

	return fmt.Sprintf("%d°%d'%.2f\"%s", degrees, minutes, seconds, direction)
}
