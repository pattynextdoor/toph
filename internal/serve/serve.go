// Package serve runs the toph dashboard as a Wish SSH server, allowing
// remote users to view agent activity over an SSH connection. Each SSH
// session gets its own Bubble Tea program backed by the shared Manager.
package serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/wish/v2"
	wishbt "charm.land/wish/v2/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/model"
)

// Run starts a Wish SSH server on the given port. It serves the toph
// dashboard to each connecting client over SSH, using public-key auth
// only (no passwords). The host key is stored at ~/.config/toph/host_key
// and generated automatically on first run.
//
// The caller is responsible for feeding events into the Manager before
// calling Run — each SSH session's Bubble Tea program reads from the
// shared Manager on its 30fps tick loop.
//
// The server blocks until ctx is cancelled or a fatal error occurs.
func Run(ctx context.Context, port int, mgr *data.Manager) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}

	keyDir := filepath.Join(home, ".config", "toph")
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	keyPath := filepath.Join(keyDir, "host_key")

	srv, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("127.0.0.1:%d", port)),
		wish.WithHostKeyPath(keyPath),
		// Accept any valid SSH key — no password auth.
		wish.WithPublicKeyAuth(func(_ ssh.Context, _ ssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			wishbt.Middleware(func(_ ssh.Session) (tea.Model, []tea.ProgramOption) {
				m := model.New(mgr)
				return m, nil
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("create SSH server: %w", err)
	}

	// Disable password auth explicitly by rejecting all passwords.
	srv.PasswordHandler = func(_ ssh.Context, _ string) bool {
		return false
	}

	fmt.Printf("toph SSH server listening on 127.0.0.1:%d\n", port)
	fmt.Printf("Connect with: ssh -p %d localhost\n", port)

	// Shut down gracefully when the context is cancelled.
	go func() {
		<-ctx.Done()
		slog.Debug("shutting down SSH server")
		_ = srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && ctx.Err() == nil {
		return fmt.Errorf("SSH server error: %w", err)
	}

	return nil
}
