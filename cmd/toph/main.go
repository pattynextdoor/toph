package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/pattynextdoor/toph/internal/config"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/model"
	"github.com/pattynextdoor/toph/internal/setup"
	"github.com/pattynextdoor/toph/internal/source"
)

// version and commit are set at build time via ldflags by goreleaser.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	cfg := config.Parse()

	if cfg.Version {
		fmt.Printf("toph %s (%s)\n", version, commit)
		return
	}

	// Handle setup subcommand: modify Claude Code settings and exit (no TUI needed)
	if cfg.Command == config.CmdSetup {
		if cfg.SetupRemove {
			if err := setup.Remove(); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing hooks: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("toph hooks removed from Claude Code settings.")
		} else {
			if err := setup.Install(7891); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing hooks: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("toph hooks installed in Claude Code settings.")
			fmt.Println("Hooks will POST to http://127.0.0.1:7891/hook")
		}
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot find home directory: %v\n", err)
		os.Exit(1)
	}

	if cfg.Debug {
		logDir := filepath.Join(home, ".config", "toph")
		os.MkdirAll(logDir, 0755)
		logFile, err := os.OpenFile(filepath.Join(logDir, "toph.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			slog.SetDefault(slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug})))
			defer logFile.Close()
		}
	}

	projectsDir := filepath.Join(home, ".claude", "projects")
	slog.Debug("toph starting", "version", version, "projects_dir", projectsDir)

	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No Claude Code data found at %s\n", projectsDir)
		fmt.Fprintf(os.Stderr, "Run Claude Code at least once to generate session data.\n")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	manager := data.NewManager()
	jsonlSource := source.NewJSONLSource(projectsDir)
	jsonlSource.SetManager(manager)
	eventCh := make(chan data.Event, 256)

	// Start JSONL source in background goroutine
	go func() {
		if err := jsonlSource.Start(ctx, eventCh); err != nil {
			fmt.Fprintf(os.Stderr, "Source error: %v\n", err)
		}
	}()

	// Start hook HTTP server for real-time Claude Code hook events
	hookSource := source.NewHookSource(0)
	go func() {
		if err := hookSource.Start(ctx, eventCh); err != nil {
			slog.Debug("hooks source error", "error", err)
		}
	}()
	slog.Debug("hooks server port", "port", hookSource.Port())

	// Start process scanner (30s interval) for supplementary session detection
	processSource := source.NewProcessSource(30 * time.Second)
	go func() {
		if err := processSource.Start(ctx, eventCh); err != nil {
			slog.Error("process source error", "err", err)
		}
	}()

	m := model.New(manager)
	p := tea.NewProgram(m)

	// Bridge source events to Bubble Tea messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-eventCh:
				if !ok {
					return
				}
				p.Send(model.EventMsg(e))
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cancel()
	jsonlSource.Stop()
	hookSource.Stop()
	processSource.Stop()
}
