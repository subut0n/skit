package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAndRecent(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		filePath: filepath.Join(dir, "history.json"),
	}

	if err := m.Add("test", "vitest run", "npm"); err != nil {
		t.Fatal(err)
	}
	if err := m.Add("build", "next build", "bun"); err != nil {
		t.Fatal(err)
	}

	entries := m.Recent(10)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Most recent first
	if entries[0].Script != "build" {
		t.Errorf("entries[0].Script = %q, want 'build'", entries[0].Script)
	}
	if entries[1].Script != "test" {
		t.Errorf("entries[1].Script = %q, want 'test'", entries[1].Script)
	}
}

func TestMaxEntries(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		filePath: filepath.Join(dir, "history.json"),
	}

	for i := 0; i < 60; i++ {
		_ = m.Add("script", "cmd", "npm")
	}

	entries := m.Recent(100)
	if len(entries) != 50 {
		t.Errorf("expected max 50 entries, got %d", len(entries))
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	m1 := &Manager{filePath: path}
	_ = m1.Add("test", "vitest", "bun")

	m2 := &Manager{filePath: path}
	_ = m2.load()

	entries := m2.Recent(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after reload, got %d", len(entries))
	}
	if entries[0].Script != "test" {
		t.Errorf("reloaded entry = %q, want 'test'", entries[0].Script)
	}
}

func TestRecentLargerThanEntries(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{
		filePath: filepath.Join(dir, "history.json"),
	}

	_ = m.Add("test", "cmd", "npm")
	entries := m.Recent(100)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestLoadNonExistent(t *testing.T) {
	m := &Manager{
		filePath: filepath.Join(os.TempDir(), "nonexistent-skit-test.json"),
	}
	err := m.load()
	if err == nil {
		t.Error("expected error loading non-existent file")
	}
}
