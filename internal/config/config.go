package config

import "os"

// Command represents the subcommand to run.
type Command int

const (
	CmdDashboard Command = iota
	CmdSetup
)

// Config holds CLI configuration for toph.
type Config struct {
	Command     Command
	Debug       bool
	Version     bool
	SetupRemove bool
}

// Parse reads CLI args and returns a populated Config.
// We use manual parsing instead of flag because flag doesn't handle subcommands well.
func Parse() *Config {
	c := &Config{}
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "setup":
			c.Command = CmdSetup
		case "--remove":
			c.SetupRemove = true
		case "--debug":
			c.Debug = true
		case "--version", "-v":
			c.Version = true
		}
	}

	return c
}
