package events

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	mutex   = &sync.Mutex{}
	events  []Event
	logFile *os.File
)

func Init() {
	// Create log file with current timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join("logs", fmt.Sprintf("events_%s.log", timestamp))

	var err error
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	// Write initial log entry
	logFile.WriteString(fmt.Sprintf("=== Event Log Started at %s ===\n", time.Now().Format("2006-01-02 15:04:05")))
}

func LogEvent(event Event) {

	mutex.Lock()
	defer mutex.Unlock()
	events = append(events, event)

	if logFile == nil {
		return
	}

	// Format: [timestamp] EVENT_TYPE: program_name
	logLine := fmt.Sprintf("[%s] %s: %s\n",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		strings.ToUpper(event.Type),
		event.Program)

	if _, err := logFile.WriteString(logLine); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}

}

// GetEvents returns the recent events (last 50)
func GetEvents() []Event {
	mutex.Lock()
	defer mutex.Unlock()
	
	start := 0
	if len(events) > 50 {
		start = len(events) - 50
	}
	return events[start:]
}
