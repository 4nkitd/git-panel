package git

import (
	"testing"
)

func TestAssignGraphColumns(t *testing.T) {
	tests := []struct {
		name         string
		commits      []CommitInfo
		checkColumns func(t *testing.T, result *LogResult)
	}{
		{
			name: "single commit",
			commits: []CommitInfo{
				{Hash: "aaa", Parents: []string{}},
			},
			checkColumns: func(t *testing.T, result *LogResult) {
				if len(result.Commits) != 1 {
					t.Fatalf("expected 1 commit, got %d", len(result.Commits))
				}
				if result.Commits[0].BranchIdx < 0 {
					t.Errorf("BranchIdx should be >= 0, got %d", result.Commits[0].BranchIdx)
				}
			},
		},
		{
			name: "linear history",
			commits: []CommitInfo{
				{Hash: "ccc", Parents: []string{"bbb"}},
				{Hash: "bbb", Parents: []string{"aaa"}},
				{Hash: "aaa", Parents: []string{}},
			},
			checkColumns: func(t *testing.T, result *LogResult) {
				for i, c := range result.Commits {
					if c.BranchIdx < 0 {
						t.Errorf("commit %d: BranchIdx should be >= 0, got %d", i, c.BranchIdx)
					}
				}
			},
		},
		{
			name: "merge commit",
			commits: []CommitInfo{
				{Hash: "merge", Parents: []string{"main", "feature"}},
				{Hash: "main", Parents: []string{"base"}},
				{Hash: "feature", Parents: []string{"base"}},
				{Hash: "base", Parents: []string{}},
			},
			checkColumns: func(t *testing.T, result *LogResult) {
				merge := result.Commits[0]
				if len(merge.Parents) != 2 {
					t.Errorf("merge commit should have 2 parents, got %d", len(merge.Parents))
				}
				if merge.BranchIdx < 0 {
					t.Errorf("merge BranchIdx should be >= 0, got %d", merge.BranchIdx)
				}
			},
		},
		{
			name:    "empty log",
			commits: []CommitInfo{},
			checkColumns: func(t *testing.T, result *LogResult) {
				if len(result.Commits) != 0 {
					t.Errorf("expected 0 commits, got %d", len(result.Commits))
				}
			},
		},
		{
			name: "many branches capped",
			commits: []CommitInfo{
				{Hash: "m1", Parents: []string{"b1", "b2", "b3", "b4", "b5", "b6", "b7"}},
				{Hash: "b1", Parents: []string{}},
				{Hash: "b2", Parents: []string{}},
				{Hash: "b3", Parents: []string{}},
				{Hash: "b4", Parents: []string{}},
				{Hash: "b5", Parents: []string{}},
				{Hash: "b6", Parents: []string{}},
				{Hash: "b7", Parents: []string{}},
			},
			checkColumns: func(t *testing.T, result *LogResult) {
				for i, c := range result.Commits {
					if c.BranchIdx > 6 {
						t.Errorf("commit %d: BranchIdx should be <= 6 (capped), got %d", i, c.BranchIdx)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &LogResult{Commits: tt.commits}
			assignGraphColumns(result)
			tt.checkColumns(t, result)
		})
	}
}

func TestCommitInfoFields(t *testing.T) {
	commit := CommitInfo{
		Hash:      "abc123def456",
		ShortHash: "abc123d",
		Subject:   "feat: add new feature",
		Author:    "John Doe",
		Date:      "2 hours ago",
		Refs:      []string{"HEAD -> main", "origin/main"},
		Parents:   []string{"parent1", "parent2"},
		BranchIdx: 0,
		Expanded:  false,
		Files:     nil,
	}

	if commit.Hash != "abc123def456" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc123def456")
	}
	if commit.ShortHash != "abc123d" {
		t.Errorf("ShortHash = %q, want %q", commit.ShortHash, "abc123d")
	}
	if len(commit.Refs) != 2 {
		t.Errorf("len(Refs) = %d, want 2", len(commit.Refs))
	}
	if len(commit.Parents) != 2 {
		t.Errorf("len(Parents) = %d, want 2", len(commit.Parents))
	}
}

func TestCommitFileFields(t *testing.T) {
	cf := CommitFile{
		Path:   "internal/app/model.go",
		Status: StatusModified,
	}

	if cf.Path != "internal/app/model.go" {
		t.Errorf("Path = %q, want %q", cf.Path, "internal/app/model.go")
	}
	if cf.Status != StatusModified {
		t.Errorf("Status = %v, want %v", cf.Status, StatusModified)
	}
}

func TestLogResultEmpty(t *testing.T) {
	result := &LogResult{}
	if len(result.Commits) != 0 {
		t.Errorf("expected empty Commits slice")
	}
}
