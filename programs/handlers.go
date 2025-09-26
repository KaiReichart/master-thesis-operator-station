package programs

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/kaireichart/master-thesis-operator-station/events"
)

//go:generate go tool templ generate

func serveProgramManager(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := ProgramManager().Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SetupHandlers() {
	http.HandleFunc("/program-manager", serveProgramManager)

	http.HandleFunc("/programs/status-all", handleStatusAll)
	http.HandleFunc("/programs/launch", handleLaunchHTMX)
	http.HandleFunc("/programs/kill", handleKillHTMX)
}

// HTMX Handlers

func handleStatusAll(w http.ResponseWriter, r *http.Request) {
	programs := GetPrograms()
	states := GetProgramStates()

	w.Header().Set("Content-Type", "text/html")
	err := ProgramList(programs, states).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLaunchHTMX(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	mutex.Lock()
	program, exists := programs[name]
	if !exists {
		mutex.Unlock()
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	state, exists := programStates[name]
	if exists && state.Running {
		mutex.Unlock()
		// Return the current card state without changes
		w.Header().Set("Content-Type", "text/html")
		err := ProgramCard(name, program, state).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	cmd := exec.Command(program.Path)
	err := cmd.Start()
	if err != nil {
		mutex.Unlock()
		log.Printf("Failed to launch %s: %v", name, err)
		http.Error(w, fmt.Sprintf("Failed to start program: %v", err), http.StatusInternalServerError)
		return
	}

	programStates[name] = &ProgramState{Running: true, Cmd: cmd}
	mutex.Unlock()

	// Create and record the event
	event := events.Event{
		Type:      "launch",
		Program:   name,
		Timestamp: time.Now(),
	}
	events.LogEvent(event)

	// Return updated program card
	w.Header().Set("Content-Type", "text/html")
	err = ProgramCard(name, program, programStates[name]).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleKillHTMX(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	mutex.Lock()
	program, exists := programs[name]
	if !exists {
		mutex.Unlock()
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	// Check if the process is actually running
	if !isAppRunning(program.Name) {
		mutex.Unlock()
		// Return current state without changes
		state := programStates[name]
		if state == nil {
			state = &ProgramState{Running: false}
		}

		w.Header().Set("Content-Type", "text/html")
		err := ProgramCard(name, program, state).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	cmd := exec.Command("taskkill", "/F", "/IM", program.Name)
	err := cmd.Run()
	if err != nil {
		mutex.Unlock()
		http.Error(w, fmt.Sprintf("Failed to kill program: %v", err), http.StatusInternalServerError)
		return
	}

	// Update the state
	if state, exists := programStates[name]; exists {
		state.Running = false
	} else {
		programStates[name] = &ProgramState{Running: false}
	}
	mutex.Unlock()

	// Create and record the event
	event := events.Event{
		Type:      "kill",
		Program:   name,
		Timestamp: time.Now(),
	}
	events.LogEvent(event)

	// Return updated program card
	w.Header().Set("Content-Type", "text/html")
	err = ProgramCard(name, program, programStates[name]).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
