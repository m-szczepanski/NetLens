// Package config centralizes runtime configuration for the NetLens binary:
// command-line flags with environment fallbacks, validated at startup so
// misconfiguration fails visibly instead of silently degrading behavior.
package config

import (
	"flag"
	"fmt"
	"net"
	"os"
)

const (
	// DefaultListenAddr is loopback-only: NetLens serves a local web UI and
	// must not expose scan results on the LAN by accident.
	DefaultListenAddr = "127.0.0.1:8080"
)

// Config is the validated runtime configuration.
type Config struct {
	// ListenAddr is the HTTP listen address for API and embedded frontend.
	ListenAddr string
	// Interface optionally overrides the network interface used for
	// discovery. Empty means auto-detect.
	Interface string
}

// Load builds a Config from command-line arguments (flag.Parse semantics),
// falling back to NETLENS_* environment variables, then to defaults.
func Load(args []string) (Config, error) {
	cfg := Config{}

	fs := flag.NewFlagSet("netlens", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&cfg.ListenAddr, "listen", envOr("NETLENS_LISTEN", DefaultListenAddr), "HTTP listen address")
	fs.StringVar(&cfg.Interface, "interface", os.Getenv("NETLENS_INTERFACE"), "network interface override for discovery (default: auto-detect)")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if _, _, err := net.SplitHostPort(cfg.ListenAddr); err != nil {
		return Config{}, fmt.Errorf("invalid listen address %q: %w", cfg.ListenAddr, err)
	}
	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
