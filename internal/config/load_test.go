package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FontSize != 14 {
		t.Errorf("expected FontSize 14, got %d", cfg.FontSize)
	}
}

func TestLoad_YamlFile(t *testing.T) {
	dir := t.TempDir()
	yaml := "fontSize: 18\n"
	if err := os.WriteFile(filepath.Join(dir, "setting.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FontSize != 18 {
		t.Errorf("expected FontSize 18, got %d", cfg.FontSize)
	}
}

func TestLoad_LocalOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "setting.yaml"), []byte("fontSize: 18\n"), 0644)
	os.WriteFile(filepath.Join(dir, "setting.local.yaml"), []byte("fontSize: 22\n"), 0644)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FontSize != 22 {
		t.Errorf("expected FontSize 22, got %d", cfg.FontSize)
	}
}
