package config

import "flag"

// Config holds CLI flags for toph.
type Config struct {
	Debug bool
}

// Parse reads CLI flags and returns a populated Config.
func Parse() *Config {
	c := &Config{}
	flag.BoolVar(&c.Debug, "debug", false, "enable debug logging to ~/.config/toph/toph.log")
	flag.Parse()
	return c
}
