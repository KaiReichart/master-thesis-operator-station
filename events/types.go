package events

import "time"

type Event struct {
	Type      string    `json:"type"`      // "launch", "kill", "failure_started", "failure_recognised", "back_on_track", "flight_started", "flight_ended", "confused"
	Program   string    `json:"program"`   // program name
	Timestamp time.Time `json:"timestamp"` // when the event occurred
}
