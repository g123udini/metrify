package service

import (
	"os"
	"path/filepath"
	"testing"
)

type testCfg struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
}

func TestFromFile_OK(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.json")

	input := `{"address":"localhost","port":8080,"enabled":true}`
	if err := os.WriteFile(p, []byte(input), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cfg, err := FromFile[testCfg](p)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected config, got nil")
	}

	if cfg.Address != "localhost" {
		t.Fatalf("Address=%q want %q", cfg.Address, "localhost")
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port=%d want %d", cfg.Port, 8080)
	}
	if cfg.Enabled != true {
		t.Fatalf("Enabled=%v want %v", cfg.Enabled, true)
	}
}

func TestFromFile_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := FromFile[testCfg]("/path/that/does/not/exist.json")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestFromFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")

	input := `{"address": "localhost",` // обрезанный JSON
	if err := os.WriteFile(p, []byte(input), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := FromFile[testCfg](p)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
