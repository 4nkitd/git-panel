package git

import (
	"strings"
	"testing"
)

func TestCharToStatus(t *testing.T) {
	tests := []struct {
		input    byte
		expected FileStatus
	}{
		{'M', StatusModified},
		{'A', StatusAdded},
		{'D', StatusDeleted},
		{'R', StatusRenamed},
		{'C', StatusCopied},
		{'T', StatusTypeChanged},
		{'?', StatusModified},
		{'X', StatusModified},
		{' ', StatusModified},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := charToStatus(tt.input)
			if result != tt.expected {
				t.Errorf("charToStatus(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseStatusEntry(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantNil  bool
		expected *StatusEntry
	}{
		{
			name:    "empty line",
			line:    "",
			wantNil: true,
		},
		{
			name:    "too few fields",
			line:    "1 XY sub mH mI mW hH hI",
			wantNil: true,
		},
		{
			name:    "short XY",
			line:    "1 M N... 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 file.txt",
			wantNil: true,
		},
		{
			name:    "no staged change (dot)",
			line:    "1 .M N... 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 file.txt",
			wantNil: true,
		},
		{
			name: "staged modification",
			line: "1 M. N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 file.txt",
			expected: &StatusEntry{
				Path:     "file.txt",
				Status:   StatusModified,
				IsStaged: true,
			},
		},
		{
			name: "staged add",
			line: "1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 abcdef0123456789abcdef0123456789abcdef01 newfile.go",
			expected: &StatusEntry{
				Path:     "newfile.go",
				Status:   StatusAdded,
				IsStaged: true,
			},
		},
		{
			name: "staged delete",
			line: "1 D. N... 100644 000000 000000 abcdef0123456789abcdef0123456789abcdef01 0000000000000000000000000000000000000000 oldfile.go",
			expected: &StatusEntry{
				Path:     "oldfile.go",
				Status:   StatusDeleted,
				IsStaged: true,
			},
		},
		{
			name: "staged copy",
			line: "1 C. N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 copy.txt",
			expected: &StatusEntry{
				Path:     "copy.txt",
				Status:   StatusCopied,
				IsStaged: true,
			},
		},
		{
			name: "staged type change",
			line: "1 T. N... 100644 100755 100755 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 symlink",
			expected: &StatusEntry{
				Path:     "symlink",
				Status:   StatusTypeChanged,
				IsStaged: true,
			},
		},
		{
			name: "rename entry type 2",
			line: "2 R. N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 R100\tnewname.txt\toldname.txt",
			expected: &StatusEntry{
				Path:     "newname.txt",
				Status:   StatusRenamed,
				IsStaged: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStatusEntry(tt.line)
			if tt.wantNil {
				if result != nil {
					t.Errorf("parseStatusEntry(%q) = %v, want nil", tt.line, result)
				}
				return
			}
			if result == nil {
				t.Fatalf("parseStatusEntry(%q) = nil, want non-nil", tt.line)
			}
			if result.Path != tt.expected.Path {
				t.Errorf("Path = %q, want %q", result.Path, tt.expected.Path)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", result.Status, tt.expected.Status)
			}
			if result.IsStaged != tt.expected.IsStaged {
				t.Errorf("IsStaged = %v, want %v", result.IsStaged, tt.expected.IsStaged)
			}
		})
	}
}

func TestParseUnstagedEntry(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantNil  bool
		expected *StatusEntry
	}{
		{
			name:    "empty line",
			line:    "",
			wantNil: true,
		},
		{
			name:    "too few fields",
			line:    "1 XY sub",
			wantNil: true,
		},
		{
			name:    "short XY",
			line:    "1 M N... 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 0000000000000000000000000000000000000000 file.txt",
			wantNil: true,
		},
		{
			name:    "no worktree change (dot)",
			line:    "1 M. N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 file.txt",
			wantNil: true,
		},
		{
			name: "unstaged modification",
			line: "1 .M N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 file.txt",
			expected: &StatusEntry{
				Path:   "file.txt",
				Status: StatusModified,
			},
		},
		{
			name: "unstaged delete",
			line: "1 .D N... 100644 100644 000000 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 deleted.txt",
			expected: &StatusEntry{
				Path:   "deleted.txt",
				Status: StatusDeleted,
			},
		},
		{
			name: "both staged and unstaged modification",
			line: "1 MM N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 both.txt",
			expected: &StatusEntry{
				Path:   "both.txt",
				Status: StatusModified,
			},
		},
		{
			name: "rename with unstaged change",
			line: "2 .R N... 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 abcdef0123456789abcdef0123456789abcdef01 R100\tnew.txt\told.txt",
			expected: &StatusEntry{
				Path:   "new.txt",
				Status: StatusRenamed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUnstagedEntry(tt.line)
			if tt.wantNil {
				if result != nil {
					t.Errorf("parseUnstagedEntry(%q) = %v, want nil", tt.line, result)
				}
				return
			}
			if result == nil {
				t.Fatalf("parseUnstagedEntry(%q) = nil, want non-nil", tt.line)
			}
			if result.Path != tt.expected.Path {
				t.Errorf("Path = %q, want %q", result.Path, tt.expected.Path)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", result.Status, tt.expected.Status)
			}
		})
	}
}

func TestParseConflictPath(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "standard unmerged entry",
			line:     "u UU N... 100644 100644 100644 100644 abcdef0123456789abcdef0123456789abcdef01 fedcba9876543210fedcba9876543210fedcba98 0123456789abcdef0123456789abcdef012 conflicted.txt",
			expected: "conflicted.txt",
		},
		{
			name:     "too few fields",
			line:     "u UU N... 0000 0000 0000 0000 0000 0000",
			expected: "",
		},
		{
			name:     "exactly 11 fields",
			line:     "u UU N... 100644 100644 100644 100644 abcdef01 fedcba98 01234567 path.go",
			expected: "path.go",
		},
		{
			name:     "extra fields - last field is path",
			line:     "u UU N... 100644 100644 100644 100644 abcdef01 fedcba98 01234567 file.txt extra",
			expected: "file.txt",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConflictPath(tt.line)
			if result != tt.expected {
				t.Errorf("parseConflictPath(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		raw          string
		wantLines    int
		checkContent func(t *testing.T, result *DiffResult)
	}{
		{
			name:      "empty diff",
			filePath:  "empty.txt",
			raw:       "",
			wantLines: 0,
		},
		{
			name:     "addition only",
			filePath: "added.txt",
			raw: `diff --git a/added.txt b/added.txt
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/added.txt
@@ -0,0 +1,3 @@
+line 1
+line 2
+line 3`,
			wantLines: 4,
			checkContent: func(t *testing.T, result *DiffResult) {
				if result.FilePath != "added.txt" {
					t.Errorf("FilePath = %q, want %q", result.FilePath, "added.txt")
				}
				addCount := 0
				for _, l := range result.Lines {
					if l.Type == DiffAdd {
						addCount++
					}
				}
				if addCount != 3 {
					t.Errorf("expected 3 additions, got %d", addCount)
				}
			},
		},
		{
			name:     "deletion only",
			filePath: "deleted.txt",
			raw: `diff --git a/deleted.txt b/deleted.txt
--- a/deleted.txt
+++ b/deleted.txt
@@ -1,2 +0,0 @@
-line 1
-line 2`,
			wantLines: 3,
			checkContent: func(t *testing.T, result *DiffResult) {
				delCount := 0
				for _, l := range result.Lines {
					if l.Type == DiffDelete {
						delCount++
					}
				}
				if delCount != 2 {
					t.Errorf("expected 2 deletions, got %d", delCount)
				}
			},
		},
		{
			name:     "mixed changes with context",
			filePath: "mixed.txt",
			raw: `diff --git a/mixed.txt b/mixed.txt
--- a/mixed.txt
+++ b/mixed.txt
@@ -1,5 +1,5 @@
 context line 1
-old line
+new line
 context line 2`,
			wantLines: 5,
			checkContent: func(t *testing.T, result *DiffResult) {
				var adds, dels, ctx, hunks int
				for _, l := range result.Lines {
					switch l.Type {
					case DiffAdd:
						adds++
					case DiffDelete:
						dels++
					case DiffContext:
						ctx++
					case DiffHunkHeader:
						hunks++
					}
				}
				if adds != 1 || dels != 1 || ctx != 2 || hunks != 1 {
					t.Errorf("got adds=%d, dels=%d, ctx=%d, hunks=%d; want 1,1,2,1", adds, dels, ctx, hunks)
				}
			},
		},
		{
			name:     "multiple hunks",
			filePath: "multi.go",
			raw: `diff --git a/multi.go b/multi.go
--- a/multi.go
+++ b/multi.go
@@ -1,3 +1,3 @@
 line 1
-old
+new
 line 2
@@ -10,3 +10,3 @@
 line 10
-old2
+new2
 line 11`,
			wantLines: 10,
			checkContent: func(t *testing.T, result *DiffResult) {
				hunkCount := 0
				for _, l := range result.Lines {
					if l.Type == DiffHunkHeader {
						hunkCount++
					}
				}
				if hunkCount != 2 {
					t.Errorf("expected 2 hunks, got %d", hunkCount)
				}
			},
		},
		{
			name:     "skip header lines",
			filePath: "test.txt",
			raw: `diff --git a/test.txt b/test.txt
index abc1234..def5678 100644
--- a/test.txt
+++ b/test.txt
@@ -1,1 +1,1 @@
-test
+test2`,
			wantLines: 3,
			checkContent: func(t *testing.T, result *DiffResult) {
				for _, l := range result.Lines {
					if strings.Contains(l.Content, "diff --git") {
						t.Error("header line should be skipped")
					}
				}
			},
		},
		{
			name:     "line number incrementing",
			filePath: "lines.txt",
			raw: `diff --git a/lines.txt b/lines.txt
--- a/lines.txt
+++ b/lines.txt
@@ -10,3 +10,4 @@
 context 10
 context 11
+added line
 context 12`,
			wantLines: 5,
			checkContent: func(t *testing.T, result *DiffResult) {
				var ctxLines []DiffLine
				for _, l := range result.Lines {
					if l.Type == DiffContext {
						ctxLines = append(ctxLines, l)
					}
				}
				if len(ctxLines) >= 2 {
					if ctxLines[1].NewNum != ctxLines[0].NewNum+1 {
						t.Errorf("line numbers should increment: %d -> %d", ctxLines[0].NewNum, ctxLines[1].NewNum)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDiff(tt.filePath, tt.raw)
			if result == nil {
				t.Fatal("parseDiff returned nil")
			}
			if result.FilePath != tt.filePath {
				t.Errorf("FilePath = %q, want %q", result.FilePath, tt.filePath)
			}
			if len(result.Lines) != tt.wantLines {
				t.Errorf("len(Lines) = %d, want %d\nLines: %+v", len(result.Lines), tt.wantLines, result.Lines)
			}
			if tt.checkContent != nil {
				tt.checkContent(t, result)
			}
		})
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOld int
		wantNew int
	}{
		{
			name:    "standard hunk",
			line:    "@@ -10,5 +20,8 @@ function name",
			wantOld: 10,
			wantNew: 20,
		},
		{
			name:    "single line change",
			line:    "@@ -1,1 +1,1 @@",
			wantOld: 1,
			wantNew: 1,
		},
		{
			name:    "large line numbers",
			line:    "@@ -1000,50 +2000,60 @@",
			wantOld: 1000,
			wantNew: 2000,
		},
		{
			name:    "zero start",
			line:    "@@ -0,0 +1,5 @@",
			wantOld: 0,
			wantNew: 1,
		},
		{
			name:    "with function context",
			line:    "@@ -15,7 +15,7 @@ func TestSomething(t *testing.T) {",
			wantOld: 15,
			wantNew: 15,
		},
		{
			name:    "no comma edge case",
			line:    "@@ -5 +5 @@",
			wantOld: 0,
			wantNew: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldNum, newNum int
			parseHunkHeader(tt.line, &oldNum, &newNum)
			if oldNum != tt.wantOld {
				t.Errorf("oldNum = %d, want %d", oldNum, tt.wantOld)
			}
			if newNum != tt.wantNew {
				t.Errorf("newNum = %d, want %d", newNum, tt.wantNew)
			}
		})
	}
}

func TestFileStatusString(t *testing.T) {
	tests := []struct {
		status   FileStatus
		expected string
	}{
		{StatusModified, "M"},
		{StatusAdded, "A"},
		{StatusDeleted, "D"},
		{StatusRenamed, "R"},
		{StatusUntracked, "?"},
		{StatusConflicted, "U"},
		{StatusCopied, "C"},
		{StatusTypeChanged, "T"},
		{FileStatus(999), " "},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
