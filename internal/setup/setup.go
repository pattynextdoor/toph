package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// hookMarker is the matcher value used to identify toph-managed hook entries.
// This lets us add/remove our hooks without disturbing user-configured ones.
const hookMarker = "toph-managed"

type hookEntry struct {
	Matcher string    `json:"matcher"`
	Hooks   []hookCmd `json:"hooks"`
}

type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// hookEvents are the Claude Code hook events toph subscribes to.
var hookEvents = []string{
	"PreToolUse",
	"PostToolUse",
	"Stop",
	"SubagentStart",
	"SubagentStop",
	"Notification",
}

// Install adds toph hook entries to Claude Code's settings.json.
// It uses the hookMarker as the matcher so we can identify and replace our entries later.
func Install(port int) error {
	settingsPath, err := settingsFilePath()
	if err != nil {
		return err
	}

	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	for _, event := range hookEvents {
		cmd := fmt.Sprintf(
			"curl -sf -X POST http://127.0.0.1:%d/hook -H 'Content-Type: application/json' -d '{\"event\":\"%s\"}' --max-time 1 2>/dev/null || true",
			port, event,
		)

		entry := hookEntry{
			Matcher: hookMarker,
			Hooks: []hookCmd{{
				Type:    "command",
				Command: cmd,
			}},
		}

		existing, _ := hooks[event].([]interface{})
		// Remove any existing toph entries before adding the new one
		var cleaned []interface{}
		for _, e := range existing {
			if m, ok := e.(map[string]interface{}); ok {
				if m["matcher"] == hookMarker {
					continue
				}
			}
			cleaned = append(cleaned, e)
		}
		cleaned = append(cleaned, entry)
		hooks[event] = cleaned
	}

	settings["hooks"] = hooks
	return writeSettings(settingsPath, settings)
}

// Remove strips all toph-managed hook entries from Claude Code's settings.json.
// Non-toph hooks are preserved. If hooks becomes empty, the key is removed entirely.
func Remove() error {
	settingsPath, err := settingsFilePath()
	if err != nil {
		return err
	}

	settings, err := readSettings(settingsPath)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		return nil
	}

	for _, event := range hookEvents {
		existing, _ := hooks[event].([]interface{})
		var cleaned []interface{}
		for _, e := range existing {
			if m, ok := e.(map[string]interface{}); ok {
				if m["matcher"] == hookMarker {
					continue
				}
			}
			cleaned = append(cleaned, e)
		}
		if len(cleaned) > 0 {
			hooks[event] = cleaned
		} else {
			delete(hooks, event)
		}
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	return writeSettings(settingsPath, settings)
}

func settingsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

func readSettings(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return settings, nil
}

func writeSettings(path string, settings map[string]interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
