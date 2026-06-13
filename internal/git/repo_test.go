package git

import (
	"testing"
)

func TestDefaultOllamaConfig(t *testing.T) {
	cfg := DefaultOllamaConfig()

	if cfg.Host == "" {
		t.Error("Host should not be empty")
	}
	if cfg.Model == "" {
		t.Error("Model should not be empty")
	}
	if cfg.Host != "http://localhost:11434" {
		t.Logf("Host = %q (may be set via env)", cfg.Host)
	}
	if cfg.Model != "gemma3:1b" {
		t.Logf("Model = %q (may be set via env)", cfg.Model)
	}
}

func TestOllamaConfigDefaults(t *testing.T) {
	cfg := OllamaConfig{
		Host:  "",
		Model: "",
	}

	if cfg.Host != "" || cfg.Model != "" {
		t.Error("empty config should have empty fields")
	}
}

func TestOllamaRequestFields(t *testing.T) {
	req := ollamaRequest{
		Model:  "llama3",
		Prompt: "test prompt",
		System: "system prompt",
		Stream: true,
	}

	if req.Model != "llama3" {
		t.Errorf("Model = %q, want %q", req.Model, "llama3")
	}
	if req.Prompt != "test prompt" {
		t.Errorf("Prompt = %q, want %q", req.Prompt, "test prompt")
	}
	if req.Stream != true {
		t.Errorf("Stream = %v, want true", req.Stream)
	}
}

func TestOllamaResponseFields(t *testing.T) {
	resp := ollamaResponse{
		Response: "test response",
		Done:     false,
		Error:    "",
	}

	if resp.Response != "test response" {
		t.Errorf("Response = %q, want %q", resp.Response, "test response")
	}
	if resp.Done != false {
		t.Errorf("Done = %v, want false", resp.Done)
	}

	respDone := ollamaResponse{
		Response: "final",
		Done:     true,
	}

	if !respDone.Done {
		t.Error("Done should be true")
	}
}

func TestOllamaResponseError(t *testing.T) {
	resp := ollamaResponse{
		Response: "",
		Done:     false,
		Error:    "model not found",
	}

	if resp.Error != "model not found" {
		t.Errorf("Error = %q, want %q", resp.Error, "model not found")
	}
}

func TestStatusEntryFields(t *testing.T) {
	entry := StatusEntry{
		Path:     "cmd/main.go",
		Status:   StatusAdded,
		OldPath:  "",
		IsStaged: true,
	}

	if entry.Path != "cmd/main.go" {
		t.Errorf("Path = %q, want %q", entry.Path, "cmd/main.go")
	}
	if entry.Status != StatusAdded {
		t.Errorf("Status = %v, want %v", entry.Status, StatusAdded)
	}
	if !entry.IsStaged {
		t.Error("IsStaged should be true")
	}
}

func TestBranchInfoFields(t *testing.T) {
	info := BranchInfo{
		Name:       "feature/test",
		IsDetached: false,
		Upstream:   "origin/feature/test",
		Ahead:      3,
		Behind:     0,
	}

	if info.Name != "feature/test" {
		t.Errorf("Name = %q, want %q", info.Name, "feature/test")
	}
	if info.IsDetached {
		t.Error("IsDetached should be false")
	}
	if info.Ahead != 3 {
		t.Errorf("Ahead = %d, want 3", info.Ahead)
	}
}

func TestBranchInfoDetached(t *testing.T) {
	info := BranchInfo{
		Name:       "abc1234",
		IsDetached: true,
		Upstream:   "",
		Ahead:      0,
		Behind:     0,
	}

	if !info.IsDetached {
		t.Error("IsDetached should be true")
	}
	if info.Upstream != "" {
		t.Errorf("Upstream should be empty for detached HEAD, got %q", info.Upstream)
	}
}

func TestStashEntryFields(t *testing.T) {
	stash := StashEntry{
		Index:   2,
		Message: "WIP: feature branch",
	}

	if stash.Index != 2 {
		t.Errorf("Index = %d, want 2", stash.Index)
	}
	if stash.Message != "WIP: feature branch" {
		t.Errorf("Message = %q, want %q", stash.Message, "WIP: feature branch")
	}
}

func TestRepoStatusFields(t *testing.T) {
	status := RepoStatus{
		Branch: BranchInfo{
			Name: "main",
		},
		Staged: []StatusEntry{
			{Path: "file1.go", Status: StatusModified, IsStaged: true},
		},
		Unstaged: []StatusEntry{
			{Path: "file2.go", Status: StatusModified, IsStaged: false},
		},
		Stashes: []StashEntry{
			{Index: 0, Message: "test stash"},
		},
		MergeHead:  false,
		RebaseHead: false,
	}

	if status.Branch.Name != "main" {
		t.Errorf("Branch.Name = %q, want %q", status.Branch.Name, "main")
	}
	if len(status.Staged) != 1 {
		t.Errorf("len(Staged) = %d, want 1", len(status.Staged))
	}
	if len(status.Unstaged) != 1 {
		t.Errorf("len(Unstaged) = %d, want 1", len(status.Unstaged))
	}
	if len(status.Stashes) != 1 {
		t.Errorf("len(Stashes) = %d, want 1", len(status.Stashes))
	}
}

func TestDiffLineFields(t *testing.T) {
	line := DiffLine{
		Type:    DiffAdd,
		Content: "new line of code",
		OldNum:  0,
		NewNum:  42,
	}

	if line.Type != DiffAdd {
		t.Errorf("Type = %v, want %v", line.Type, DiffAdd)
	}
	if line.Content != "new line of code" {
		t.Errorf("Content = %q, want %q", line.Content, "new line of code")
	}
	if line.NewNum != 42 {
		t.Errorf("NewNum = %d, want 42", line.NewNum)
	}
}

func TestDiffLineTypes(t *testing.T) {
	types := []struct {
		lt   DiffLineType
		name string
	}{
		{DiffContext, "context"},
		{DiffAdd, "add"},
		{DiffDelete, "delete"},
		{DiffHunkHeader, "hunk header"},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			line := DiffLine{Type: tt.lt}
			if line.Type != tt.lt {
				t.Errorf("Type mismatch")
			}
		})
	}
}

func TestDiffResultFields(t *testing.T) {
	result := DiffResult{
		FilePath: "test.go",
		Lines: []DiffLine{
			{Type: DiffHunkHeader, Content: "@@ -1,3 +1,3 @@"},
			{Type: DiffContext, Content: "line 1", OldNum: 1, NewNum: 1},
			{Type: DiffDelete, Content: "old line 2", OldNum: 2},
			{Type: DiffAdd, Content: "new line 2", NewNum: 2},
		},
	}

	if result.FilePath != "test.go" {
		t.Errorf("FilePath = %q, want %q", result.FilePath, "test.go")
	}
	if len(result.Lines) != 4 {
		t.Errorf("len(Lines) = %d, want 4", len(result.Lines))
	}
}

func TestRepoPath(t *testing.T) {
	repo := &Repo{Path: "/path/to/repo"}
	if repo.Path != "/path/to/repo" {
		t.Errorf("Path = %q, want %q", repo.Path, "/path/to/repo")
	}
}
