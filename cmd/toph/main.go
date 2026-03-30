package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/model"
	"github.com/pattynextdoor/toph/internal/source"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot find home directory: %v\n", err)
		os.Exit(1)
	}
	projectsDir := filepath.Join(home, ".claude", "projects")

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
	eventCh := make(chan data.Event, 256)

	// Start JSONL source in background goroutine
	go func() {
		if err := jsonlSource.Start(ctx, eventCh); err != nil {
			fmt.Fprintf(os.Stderr, "Source error: %v\n", err)
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
}
