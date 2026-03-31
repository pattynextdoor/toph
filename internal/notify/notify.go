package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// Notifier sends desktop notifications with debouncing to avoid spam.
// Currently macOS-only (uses osascript); other platforms are silently skipped.
type Notifier struct {
	mu       sync.Mutex
	lastSent map[string]time.Time // session ID → last notification time
	cooldown time.Duration
}

// New creates a Notifier with the given cooldown between repeat notifications
// for the same session.
func New(cooldown time.Duration) *Notifier {
	return &Notifier{
		lastSent: make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// SessionWaiting sends a notification that a session is waiting for permission.
// Debounced per session: won't re-notify for the same session within the
// cooldown period.
func (n *Notifier) SessionWaiting(sessionID, project string) {
	if runtime.GOOS != "darwin" {
		return // Only macOS for now
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if last, ok := n.lastSent[sessionID]; ok {
		if time.Since(last) < n.cooldown {
			return
		}
	}

	n.lastSent[sessionID] = time.Now()

	title := "toph"
	body := fmt.Sprintf("%s is waiting for permission", project)
	if project == "" {
		body = fmt.Sprintf("Session %s is waiting for permission", sessionID[:min(8, len(sessionID))])
	}

	// Fire and forget — don't block the UI.
	go func() {
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		exec.Command("osascript", "-e", script).Run() //nolint:errcheck
	}()
}
