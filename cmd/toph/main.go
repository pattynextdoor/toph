package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/pattynextdoor/toph/internal/config"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/export"
	"github.com/pattynextdoor/toph/internal/model"
	"github.com/pattynextdoor/toph/internal/serve"
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

	// Export subcommand: read JSONL files, aggregate state, print JSON, exit.
	if cfg.Command == config.CmdExport {
		manager := data.NewManager()
		matches, _ := filepath.Glob(filepath.Join(projectsDir, "*", "*.jsonl"))
		for _, path := range matches {
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			rel, _ := filepath.Rel(projectsDir, path)
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			project := ""
			if len(parts) > 0 {
				project = parts[0]
			}
			for _, e := range source.ParseBytes(content, project) {
				manager.HandleEvent(e)
			}
		}
		if err := export.Run(manager); err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			os.Exit(1)
		}
		return
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

	// Serve mode: start an SSH server instead of a local TUI. Events flow
	// into the shared Manager; each SSH session creates its own Bubble Tea
	// program that reads from the Manager on tick.
	if cfg.Command == config.CmdServe {
		port := cfg.ServePort
		if port == 0 {
			port = 2222
		}

		// Bridge source events into the Manager so SSH sessions see live data.
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-eventCh:
					if !ok {
						return
					}
					manager.HandleEvent(e)
				}
			}
		}()

		if err := serve.Run(ctx, port, manager); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		jsonlSource.Stop()
		hookSource.Stop()
		return
	}

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
}
