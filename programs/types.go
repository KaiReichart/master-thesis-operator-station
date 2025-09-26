package programs

import "os/exec"

type Program struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	CanKill bool   `json:"canKill"`
}

type ProgramState struct {
	Running bool      `json:"running"`
	Cmd     *exec.Cmd `json:"-"`
}
