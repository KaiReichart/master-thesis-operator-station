package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

//go:generate go tool templ generate

func SetupHandlers() {
	http.HandleFunc("/events", handleEvents)
	http.HandleFunc("/manual-event", handleManualEvent)

	// New HTMX endpoints
	http.HandleFunc("/events/list", handleEventsList)
	http.HandleFunc("/events/manual", handleManualEventHTMX)
}

// HTMX Handlers

func handleEventsList(w http.ResponseWriter, r *http.Request) {
	eventsList := GetEvents()

	// Reverse the events to show newest first
	reversed := make([]Event, len(eventsList))
	for i, j := 0, len(eventsList)-1; i < len(eventsList); i, j = i+1, j-1 {
		reversed[i] = eventsList[j]
	}

	w.Header().Set("Content-Type", "text/html")
	err := EventsList(reversed).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleManualEventHTMX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventType := r.FormValue("type")
	program := r.FormValue("program")

	if eventType == "" || program == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create and record the event
	event := Event{
		Type:      eventType,
		Program:   program,
		Timestamp: time.Now(),
	}

	// Log the event to file
	LogEvent(event)

	// Return updated events list
	eventsList := GetEvents()

	// Reverse the events to show newest first
	reversed := make([]Event, len(eventsList))
	for i, j := 0, len(eventsList)-1; i < len(eventsList); i, j = i+1, j-1 {
		reversed[i] = eventsList[j]
	}

	w.Header().Set("Content-Type", "text/html")
	err := EventsList(reversed).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Legacy JSON API handlers (keeping for backward compatibility)

func handleEvents(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	// Return the last 50 events
	start := 0
	if len(events) > 50 {
		start = len(events) - 50
	}
	json.NewEncoder(w).Encode(events[start:])
}

func handleManualEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Type    string `json:"type"`
		Program string `json:"program"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create and record the event
	event := Event{
		Type:      data.Type,
		Program:   data.Program,
		Timestamp: time.Now(),
	}

	// Log the event to file
	LogEvent(event)

	w.WriteHeader(http.StatusOK)
}

// Helper functions for templates

func formatEventType(eventType string) string {
	parts := strings.Split(eventType, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func getEventTypeClass(eventType string) string {
	switch eventType {
	case "launch":
		return "bg-green-100 text-green-800"
	case "kill":
		return "bg-red-100 text-red-800"
	case "flight_started":
		return "bg-green-100 text-green-800"
	case "flight_ended":
		return "bg-red-100 text-red-800"
	case "failure_started":
		return "bg-orange-100 text-orange-800"
	case "failure_recognised":
		return "bg-purple-100 text-purple-800"
	case "confused":
		return "bg-yellow-100 text-yellow-800"
	case "preparations_started", "preparations_finished":
		return "bg-indigo-100 text-indigo-800"
	default:
		return "bg-blue-100 text-blue-800"
	}
}
