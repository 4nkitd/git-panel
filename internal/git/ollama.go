package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OllamaConfig holds configuration for the Ollama integration.
type OllamaConfig struct {
	Host  string // default: http://localhost:11434
	Model string // default: llama3
}

// DefaultOllamaConfig returns the default Ollama configuration,
// respecting OLLAMA_HOST and OLLAMA_MODEL env vars.
func DefaultOllamaConfig() OllamaConfig {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3"
	}
	return OllamaConfig{Host: host, Model: model}
}

// ollamaRequest is the JSON body for /api/generate.
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

// ollamaResponse is a single JSON line from /api/generate.
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// GenerateCommitMessage calls Ollama to generate a commit message from a diff.
// It streams partial results through the onChunk callback so the UI can update live.
// Returns the full generated message.
func GenerateCommitMessage(ctx context.Context, cfg OllamaConfig, diff string, onChunk func(partial string)) (string, error) {
	// Truncate very large diffs to avoid overwhelming the model
	maxDiffLen := 8000
	if len(diff) > maxDiffLen {
		diff = diff[:maxDiffLen] + "\n... (diff truncated)"
	}

	systemPrompt := `You are a commit message generator. Given a git diff, write a concise, conventional commit message.

Rules:
- Use conventional commit format: type(scope): description
- Types: feat, fix, refactor, docs, style, test, chore, perf, ci, build
- Keep the subject line under 72 characters
- Be specific about what changed, not how
- Do NOT include a body or footer unless the change is complex
- Output ONLY the commit message, nothing else — no explanation, no quotes, no markdown`

	prompt := fmt.Sprintf("Generate a commit message for this diff:\n\n```diff\n%s\n```", diff)

	reqBody := ollamaRequest{
		Model:  cfg.Model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(cfg.Host, "/") + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(body))
	}

	// Stream NDJSON response
	var full strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk ollamaResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return full.String(), fmt.Errorf("decode response: %w", err)
		}

		if chunk.Error != "" {
			return full.String(), fmt.Errorf("ollama error: %s", chunk.Error)
		}

		full.WriteString(chunk.Response)
		if onChunk != nil {
			onChunk(full.String())
		}

		if chunk.Done {
			break
		}
	}

	// Clean up the result — strip quotes, backticks, trailing whitespace
	result := strings.TrimSpace(full.String())
	result = strings.Trim(result, "`\"'")
	result = strings.TrimSpace(result)

	return result, nil
}

// GetStagedDiff returns the diff of all staged changes (for commit message generation).
func (r *Repo) GetStagedDiff() (string, error) {
	out, err := runGit(r.Path, "diff", "--cached", "--no-color")
	if err != nil {
		return "", err
	}
	return out, nil
}

// ListOllamaModels queries Ollama for available models.
func ListOllamaModels(cfg OllamaConfig) ([]string, error) {
	url := strings.TrimRight(cfg.Host, "/") + "/api/tags"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("cannot reach Ollama: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range result.Models {
		names = append(names, m.Name)
	}
	return names, nil
}
