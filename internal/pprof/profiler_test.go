package pprof

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCPUCapture(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "cpu.prof")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CPUCapture(ctx, filename, 1*time.Second)
	if err != nil {
		t.Fatalf("CPUCapture failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Expected profile file %s to be created", filename)
	}
}

func TestCPUCapture_ContextCancelled(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "cpu_cancel.prof")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CPUCapture(ctx, filename, 1*time.Second)
	if err == nil {
		t.Fatalf("Expected error due to context cancellation, but got none")
	}
}

func TestCapture(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "heap.prof")

	err := Capture(Heap, filename)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Expected heap profile file %s to be created", filename)
	}
}

func TestCapture_InvalidProfile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "invalid.prof")

	err := Capture("invalid_profile", filename)
	if err == nil {
		t.Fatalf("Expected error for invalid profile, but got none")
	}

	// Убедимся, что файл не был создан
	if _, err := os.Stat(filename); err == nil {
		t.Fatalf("Profile file should not have been created")
	}
}

func TestCapture_FileCreationFailure(t *testing.T) {
	filename := filepath.Join("/invalid_path", "heap.prof")

	err := Capture(Heap, filename)
	if err == nil {
		t.Fatalf("Expected error for invalid file path, but got none")
	}
}

func TestCPUCapture_FileCreationFailure(t *testing.T) {
	filename := filepath.Join("/invalid_path", "cpu.prof")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CPUCapture(ctx, filename, 1*time.Second)
	if err == nil {
		t.Fatalf("Expected error for invalid file path, but got none")
	}
}
