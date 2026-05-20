package permissions

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Action string

const (
	ReadFile   Action = "read_file"
	WriteFile  Action = "write_file"
	ListDir    Action = "list_dir"
	RunCommand Action = "run_command"
)

type Decision int

const (
	Denied Decision = iota
	Allowed
	AlwaysAllowed
)

type Manager struct {
	always map[Action]bool
	denied map[Action]int
	UI     UI
}

type UI interface {
	RenderPermissionRequest(action Action, detail string)
	RenderPermissionApproved(action Action, detail string)
	RenderPermissionDenied()
}

func NewManager() *Manager {
	return &Manager{
		always: map[Action]bool{},
		denied: map[Action]int{},
	}
}

func (m *Manager) IsAlwaysAllowed(action Action) bool {
	return m.always[action]
}

func (m *Manager) DeniedCount(action Action) int {
	return m.denied[action]
}

func (m *Manager) Status(action Action) string {
	if m.always[action] {
		return "allowed"
	}
	return "ask"
}

func (m *Manager) Approve(action Action, detail string, in io.Reader, out io.Writer) (bool, error) {
	if m.always[action] {
		return true, nil
	}

	if m.UI != nil {
		m.UI.RenderPermissionRequest(action, detail)
	} else {
		fmt.Fprintf(out, "\nShard wants to:\n%s: %s\n\nAllow?\n[y] yes  [n] no  [a] always for this session\n> ", action, detail)
	}
	if m.UI != nil {
		fmt.Fprint(out, "> ")
	}
	reader := bufio.NewReader(in)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		if m.UI != nil {
			m.UI.RenderPermissionApproved(action, detail)
		}
		return true, nil
	case "a", "always":
		m.always[action] = true
		if m.UI != nil {
			m.UI.RenderPermissionApproved(action, detail)
		}
		return true, nil
	default:
		m.denied[action]++
		if m.UI != nil {
			m.UI.RenderPermissionDenied()
		}
		return false, nil
	}
}

func Actions() []Action {
	return []Action{ReadFile, WriteFile, ListDir, RunCommand}
}
