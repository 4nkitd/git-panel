package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if cfg.Ollama.Host == "" {
		t.Error("Ollama.Host should not be empty")
	}
	if cfg.Ollama.Model == "" {
		t.Error("Ollama.Model should not be empty")
	}
	if cfg.UI.GraphMaxVisible <= 0 {
		t.Error("GraphMaxVisible should be positive")
	}
	if cfg.Editor.Command == "" {
		t.Error("Editor.Command should not be empty")
	}
}

func TestDefaultValues(t *testing.T) {
	cfg := Default()

	if cfg.UI.GraphCollapsed != false {
		t.Errorf("GraphCollapsed = %v, want false", cfg.UI.GraphCollapsed)
	}
	if cfg.UI.StashCollapsed != true {
		t.Errorf("StashCollapsed = %v, want true", cfg.UI.StashCollapsed)
	}
	if cfg.UI.ShowUntracked != true {
		t.Errorf("ShowUntracked = %v, want true", cfg.UI.ShowUntracked)
	}
	if cfg.UI.ConfirmDiscard != true {
		t.Errorf("ConfirmDiscard = %v, want true", cfg.UI.ConfirmDiscard)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	val := getEnvOrDefault("NONEXISTENT_VAR_12345", "default")
	if val != "default" {
		t.Errorf("getEnvOrDefault for nonexistent var = %q, want 'default'", val)
	}

	os.Setenv("TEST_CONFIG_VAR", "test_value")
	val = getEnvOrDefault("TEST_CONFIG_VAR", "default")
	if val != "test_value" {
		t.Errorf("getEnvOrDefault = %q, want 'test_value'", val)
	}
	os.Unsetenv("TEST_CONFIG_VAR")
}

func TestConfigPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Error("Path() should not return empty string")
	}
	if filepath.Base(path) != ".git-panel.yaml" {
		t.Errorf("Path() base = %q, want '.git-panel.yaml'", filepath.Base(path))
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil for non-existent config")
	}
	if cfg.Ollama.Host == "" {
		t.Error("Load() should return valid defaults")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-panel-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	homeBackup := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", homeBackup)

	cfg := Default()
	cfg.Ollama.Model = "test-model"
	cfg.UI.GraphCollapsed = true

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	loaded := Load()
	if loaded.Ollama.Model != "test-model" {
		t.Errorf("Loaded model = %q, want 'test-model'", loaded.Ollama.Model)
	}
	if !loaded.UI.GraphCollapsed {
		t.Error("Loaded GraphCollapsed should be true")
	}
}

func TestSaveOllama(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "git-panel-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	homeBackup := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", homeBackup)

	cfg := Default()
	if err := cfg.SaveOllama("http://test:1234", "custom-model"); err != nil {
		t.Fatalf("SaveOllama() failed: %v", err)
	}

	loaded := Load()
	if loaded.Ollama.Host != "http://test:1234" {
		t.Errorf("Host = %q, want 'http://test:1234'", loaded.Ollama.Host)
	}
	if loaded.Ollama.Model != "custom-model" {
		t.Errorf("Model = %q, want 'custom-model'", loaded.Ollama.Model)
	}
}

func TestGetEditorCommand(t *testing.T) {
	edBackup := os.Getenv("EDITOR")
	visBackup := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", edBackup)
		os.Setenv("VISUAL", visBackup)
	}()

	os.Unsetenv("EDITOR")
	os.Unsetenv("VISUAL")
	cmd := getEditorCommand()
	if cmd != "vim" {
		t.Errorf("getEditorCommand() = %q, want 'vim' (default)", cmd)
	}

	os.Setenv("EDITOR", "nano")
	cmd = getEditorCommand()
	if cmd != "nano" {
		t.Errorf("getEditorCommand() = %q, want 'nano'", cmd)
	}

	os.Setenv("VISUAL", "code")
	os.Unsetenv("EDITOR")
	cmd = getEditorCommand()
	if cmd != "code" {
		t.Errorf("getEditorCommand() = %q, want 'code' (from VISUAL)", cmd)
	}
}
