package programs

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	programs      = map[string]Program{}
	programStates = map[string]*ProgramState{}
	mutex         = &sync.Mutex{}
)

func Init() {
	programs["FS2FF"] = Program{
		Name:    "fs2ff.exe",
		Path:    "C:\\Users\\kai\\Documents\\fs2ff.exe",
		CanKill: true,
	}
	programs["SkyDolly"] = Program{
		Name:    "SkyDolly.exe",
		Path:    "C:\\Users\\kai\\Documents\\SkyDolly\\SkyDolly.exe",
		CanKill: false,
	}
	programs["FS-FlightControl"] = Program{
		Name:    "FS-FlightControl.exe",
		Path:    "C:\\Program Files\\FS-FlightControl\\FS-FlightControl.exe",
		CanKill: false,
	}

	// Initialize program states
	for name := range programs {
		programStates[name] = &ProgramState{Running: isAppRunning(programs[name].Name)}
	}
	go monitorProgramStates()
}

func isAppRunning(name string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", name))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), name)
}

func monitorProgramStates() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for name, state := range programStates {
			if state.Running {
				state.Running = isAppRunning(name)
			}
		}
		mutex.Unlock()
	}
}

// GetPrograms returns a copy of the programs map
func GetPrograms() map[string]Program {
	return programs
}

// GetProgramStates returns a copy of the program states map
func GetProgramStates() map[string]*ProgramState {
	mutex.Lock()
	defer mutex.Unlock()
	
	// Update states before returning
	for name, program := range programs {
		running := isAppRunning(program.Name)
		if state, exists := programStates[name]; exists {
			state.Running = running
		} else {
			programStates[name] = &ProgramState{Running: running}
		}
	}
	
	return programStates
}
