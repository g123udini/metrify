package main

import (
	"flag"
	"os"
	"testing"
)

func resetFlags(t *testing.T) {
	t.Helper()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestSetDefaults(t *testing.T) {
	f := &flags{}
	setDefaults(f)

	if f.Host != ":8080" {
		t.Fatalf("expected host :8080, got %s", f.Host)
	}

	if f.PollInterval != 2 {
		t.Fatalf("expected poll interval 2, got %d", f.PollInterval)
	}

	if f.ReportInterval != 10 {
		t.Fatalf("expected report interval 10, got %d", f.ReportInterval)
	}

	if f.RateLimit != 1 {
		t.Fatalf("expected rate limit 1, got %d", f.RateLimit)
	}
}

func TestGetConfigPath_FromFlag(t *testing.T) {
	resetFlags(t)

	t.Setenv("CONFIG", "test-config")

	os.Args = []string{"cmd", "-config=test.json"}

	cfg := getConfigPath()

	if cfg != "test.json" {
		t.Fatalf("expected config path test.json, got %s", cfg)
	}
}

func TestGetConfigPath_FromEnv(t *testing.T) {
	resetFlags(t)

	t.Setenv("CONFIG", "env.json")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	cfg := getConfigPath()

	if cfg != "env.json" {
		t.Fatalf("expected env.json, got %s", cfg)
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags(t)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	os.Clearenv()

	f := parseFlags()

	if f.Host != ":8080" {
		t.Fatalf("expected :8080 got %s", f.Host)
	}

	if f.PollInterval != 2 {
		t.Fatalf("expected poll interval 2 got %d", f.PollInterval)
	}

	if f.ReportInterval != 10 {
		t.Fatalf("expected report interval 10 got %d", f.ReportInterval)
	}
}

func TestParseFlags_FromCLI(t *testing.T) {
	resetFlags(t)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"cmd",
		"-a", "localhost:9090",
		"-p", "5",
		"-r", "20",
		"-b",
		"-l", "10",
	}

	os.Clearenv()

	f := parseFlags()

	if f.Host != "localhost:9090" {
		t.Fatalf("expected localhost:9090 got %s", f.Host)
	}

	if f.PollInterval != 5 {
		t.Fatalf("expected 5 got %d", f.PollInterval)
	}

	if f.ReportInterval != 20 {
		t.Fatalf("expected 20 got %d", f.ReportInterval)
	}

	if !f.BatchUpdate {
		t.Fatalf("expected batch update true")
	}

	if f.RateLimit != 10 {
		t.Fatalf("expected rate limit 10 got %d", f.RateLimit)
	}
}
