package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectBunLockb(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != Bun {
		t.Errorf("expected Bun, got %d", info.Manager)
	}
	if info.RunCmd != "bun run" {
		t.Errorf("expected 'bun run', got %q", info.RunCmd)
	}
}

func TestDetectBunLock(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bun.lock"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != Bun {
		t.Errorf("expected Bun, got %d", info.Manager)
	}
}

func TestDetectPNPM(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != PNPM {
		t.Errorf("expected PNPM, got %d", info.Manager)
	}
	if info.RunCmd != "pnpm run" {
		t.Errorf("expected 'pnpm run', got %q", info.RunCmd)
	}
}

func TestDetectYarn(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != Yarn {
		t.Errorf("expected Yarn, got %d", info.Manager)
	}
	if info.RunCmd != "yarn run" {
		t.Errorf("expected 'yarn run', got %q", info.RunCmd)
	}
}

func TestDetectNPM(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != NPM {
		t.Errorf("expected NPM, got %d", info.Manager)
	}
	if info.RunCmd != "npm run" {
		t.Errorf("expected 'npm run', got %q", info.RunCmd)
	}
}

func TestDetectDefault(t *testing.T) {
	dir := t.TempDir()

	info := Detect(dir)
	if info.Manager != NPM {
		t.Errorf("expected NPM (default), got %d", info.Manager)
	}
}

func TestDetectPriority(t *testing.T) {
	dir := t.TempDir()
	// Create multiple lockfiles — bun should win
	os.WriteFile(filepath.Join(dir, "bun.lockb"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "yarn.lock"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte{}, 0644)

	info := Detect(dir)
	if info.Manager != Bun {
		t.Errorf("expected Bun (highest priority), got %d", info.Manager)
	}
}
