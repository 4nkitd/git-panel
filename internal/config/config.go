package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Ollama OllamaConfig `yaml:"ollama"`
	UI     UIConfig     `yaml:"ui"`
	Editor EditorConfig `yaml:"editor"`
}

type OllamaConfig struct {
	Host  string `yaml:"host"`
	Model string `yaml:"model"`
}

type UIConfig struct {
	GraphCollapsed    bool `yaml:"graphCollapsed"`
	StashCollapsed    bool `yaml:"stashCollapsed"`
	GraphMaxVisible   int  `yaml:"graphMaxVisible"`
	ShowUntracked     bool `yaml:"showUntracked"`
	ConfirmDiscard    bool `yaml:"confirmDiscard"`
	ConfirmLargeStage bool `yaml:"confirmLargeStage"`
}

type EditorConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

func Default() *Config {
	return &Config{
		Ollama: OllamaConfig{
			Host:  getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434"),
			Model: getEnvOrDefault("OLLAMA_MODEL", "gemma3:1b"),
		},
		UI: UIConfig{
			GraphCollapsed:    false,
			StashCollapsed:    true,
			GraphMaxVisible:   20,
			ShowUntracked:     true,
			ConfirmDiscard:    true,
			ConfirmLargeStage: false,
		},
		Editor: EditorConfig{
			Command: getEditorCommand(),
			Args:    []string{},
		},
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEditorCommand() string {
	if ed := os.Getenv("EDITOR"); ed != "" {
		return ed
	}
	if vis := os.Getenv("VISUAL"); vis != "" {
		return vis
	}
	return "vim"
}

func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".git-panel.yaml")
}

func Load() *Config {
	cfg := Default()

	path := Path()
	if path == "" {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg
	}

	return cfg
}

func (c *Config) Save() error {
	path := Path()
	if path == "" {
		return nil
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) SaveOllama(host, model string) error {
	c.Ollama.Host = host
	c.Ollama.Model = model
	return c.Save()
}
