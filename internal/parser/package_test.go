package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBasic(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")
	content := `{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "test": "vitest run"
  }
}`
	if err := os.WriteFile(pkg, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scripts, err := Parse(pkg)
	if err != nil {
		t.Fatal(err)
	}

	if len(scripts) != 3 {
		t.Fatalf("expected 3 scripts, got %d", len(scripts))
	}

	// All ungrouped, should be sorted alphabetically
	expected := []string{"build", "dev", "test"}
	for i, name := range expected {
		if scripts[i].Name != name {
			t.Errorf("scripts[%d].Name = %q, want %q", i, scripts[i].Name, name)
		}
	}
}

func TestParseWithGroups(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")
	content := `{
  "scripts": {
    "dev": "next dev",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix",
    "db:migrate": "prisma migrate deploy"
  }
}`
	if err := os.WriteFile(pkg, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scripts, err := Parse(pkg)
	if err != nil {
		t.Fatal(err)
	}

	if len(scripts) != 6 {
		t.Fatalf("expected 6 scripts, got %d", len(scripts))
	}

	// Ungrouped first (dev), then by group (db, lint, test)
	expectedOrder := []string{"dev", "db:migrate", "lint", "lint:fix", "test", "test:watch"}
	for i, name := range expectedOrder {
		if scripts[i].Name != name {
			t.Errorf("scripts[%d].Name = %q, want %q", i, scripts[i].Name, name)
		}
	}
}

func TestParseWithXSkit(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")
	content := `{
  "scripts": {
    "dev": "next dev",
    "build": "next build"
  },
  "x-skit": {
    "dev": "Start dev server with hot-reload",
    "build": "Production build"
  }
}`
	if err := os.WriteFile(pkg, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scripts, err := Parse(pkg)
	if err != nil {
		t.Fatal(err)
	}

	if len(scripts) != 2 {
		t.Fatalf("expected 2 scripts, got %d", len(scripts))
	}

	for _, s := range scripts {
		if s.Description == "" {
			t.Errorf("script %q should have a description from x-skit", s.Name)
		}
	}
}

func TestParseNoScripts(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")
	content := `{"name": "test", "version": "1.0.0"}`
	if err := os.WriteFile(pkg, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scripts, err := Parse(pkg)
	if err != nil {
		t.Fatal(err)
	}

	if scripts != nil {
		t.Fatalf("expected nil scripts, got %d", len(scripts))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkg, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Parse(pkg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFindPackageJSON(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "src", "components")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	pkgPath := filepath.Join(dir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"scripts":{}}`), 0644); err != nil {
		t.Fatal(err)
	}

	found := FindPackageJSON(sub)
	if found != pkgPath {
		t.Errorf("FindPackageJSON(%q) = %q, want %q", sub, found, pkgPath)
	}
}

func TestFindPackageJSONNotFound(t *testing.T) {
	dir := t.TempDir()
	found := FindPackageJSON(dir)
	if found != "" {
		t.Errorf("FindPackageJSON(%q) = %q, want empty", dir, found)
	}
}
