package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ListenAddr != DefaultListenAddr {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, DefaultListenAddr)
	}
	if cfg.Interface != "" {
		t.Errorf("Interface = %q, want empty (auto-detect)", cfg.Interface)
	}
}

func TestLoadFlags(t *testing.T) {
	cfg, err := Load([]string{"-listen", ":9090", "-interface", "en0"})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":9090")
	}
	if cfg.Interface != "en0" {
		t.Errorf("Interface = %q, want %q", cfg.Interface, "en0")
	}
}

func TestLoadEnvFallback(t *testing.T) {
	t.Setenv("NETLENS_LISTEN", "127.0.0.1:9999")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ListenAddr != "127.0.0.1:9999" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, "127.0.0.1:9999")
	}
}

func TestLoadRejectsInvalidListenAddr(t *testing.T) {
	if _, err := Load([]string{"-listen", "not-an-address"}); err == nil {
		t.Fatal("Load() error = nil, want invalid-address error")
	}
}
