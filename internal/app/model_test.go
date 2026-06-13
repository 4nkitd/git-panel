package app

import (
	"testing"

	"github.com/4nkitd/git-panel/internal/git"
)

func TestNewModel(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}
	if m.repo == nil {
		t.Error("repo should not be nil")
	}
	if m.mode != ModeNormal {
		t.Errorf("mode = %v, want ModeNormal", m.mode)
	}
	if m.focus != SectionUnstaged {
		t.Errorf("focus = %v, want SectionUnstaged", m.focus)
	}
	if m.commitInput.Placeholder == "" {
		t.Error("commitInput placeholder should not be empty")
	}
}

func TestModelDefaults(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	if m.stashCollapsed != true {
		t.Error("stashCollapsed should default to true")
	}
	if m.graphCollapsed != false {
		t.Error("graphCollapsed should default to false")
	}
	if m.graphMaxVisible != 20 {
		t.Errorf("graphMaxVisible = %d, want 20", m.graphMaxVisible)
	}
	if m.hoverRow != -1 {
		t.Errorf("hoverRow = %d, want -1", m.hoverRow)
	}
}

func TestSpinnerTick(t *testing.T) {
	s := Spinner{frame: 0, active: true}
	s.Tick()
	if s.frame != 1 {
		t.Errorf("frame after tick = %d, want 1", s.frame)
	}

	s.frame = len(spinnerFrames) - 1
	s.Tick()
	if s.frame != 0 {
		t.Errorf("frame should wrap to 0, got %d", s.frame)
	}
}

func TestSpinnerView(t *testing.T) {
	s := Spinner{active: false}
	if s.View() != "" {
		t.Errorf("inactive spinner should return empty string, got %q", s.View())
	}

	s.active = true
	s.frame = 0
	view := s.View()
	if view == "" {
		t.Error("active spinner should return non-empty string")
	}
}

func TestSpinnerAIView(t *testing.T) {
	s := Spinner{active: false}
	if s.AIView() != "" {
		t.Errorf("inactive spinner should return empty string, got %q", s.AIView())
	}

	s.active = true
	s.frame = 0
	view := s.AIView()
	if view == "" {
		t.Error("active spinner should return non-empty string")
	}
}

func TestCycleSection(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	sections := []Section{SectionStaged, SectionUnstaged, SectionStashes, SectionGraph}

	m.focus = SectionStaged
	for i := 0; i < len(sections)*2; i++ {
		expectedNext := sections[(i+1)%len(sections)]
		m.cycleSection(true)
		if m.focus != expectedNext {
			t.Errorf("after cycleSection(true) from %v: focus = %v, want %v", sections[i%len(sections)], m.focus, expectedNext)
		}
	}

	m.focus = SectionGraph
	for i := 0; i < len(sections)*2; i++ {
		expectedPrev := sections[(len(sections)-1-i%len(sections))%len(sections)]
		m.focus = sections[(len(sections)-i%len(sections))%len(sections)]
		m.cycleSection(false)
		if m.focus != expectedPrev {
			t.Errorf("after cycleSection(false): focus = %v, want %v", m.focus, expectedPrev)
		}
	}
}

func TestToggleCollapse(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	m.focus = SectionStaged
	m.stagedCollapsed = false
	m.toggleCollapse()
	if !m.stagedCollapsed {
		t.Error("stagedCollapsed should be true after toggle")
	}

	m.focus = SectionUnstaged
	m.unstagedCollapsed = false
	m.toggleCollapse()
	if !m.unstagedCollapsed {
		t.Error("unstagedCollapsed should be true after toggle")
	}

	m.focus = SectionStashes
	m.stashCollapsed = false
	m.toggleCollapse()
	if !m.stashCollapsed {
		t.Error("stashCollapsed should be true after toggle")
	}

	m.focus = SectionGraph
	m.graphCollapsed = false
	m.toggleCollapse()
	if !m.graphCollapsed {
		t.Error("graphCollapsed should be true after toggle")
	}
}

func TestClampCursor(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	m.status = nil
	m.clampCursor()

	m.status = &git.RepoStatus{
		Staged:   []git.StatusEntry{{Path: "file1"}},
		Unstaged: []git.StatusEntry{{Path: "file2"}, {Path: "file3"}},
	}

	m.focus = SectionUnstaged
	m.cursor = 10
	m.clampCursor()
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (len-1)", m.cursor)
	}

	m.cursor = -5
	m.clampCursor()
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}

	m.focus = SectionStaged
	m.cursor = 0
	m.clampCursor()
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestFocusedEntries(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	if entries := m.focusedEntries(); entries != nil {
		t.Errorf("focusedEntries() with nil status = %v, want nil", entries)
	}

	m.status = &git.RepoStatus{
		Staged:   []git.StatusEntry{{Path: "staged"}},
		Unstaged: []git.StatusEntry{{Path: "unstaged"}},
	}

	m.focus = SectionStaged
	entries := m.focusedEntries()
	if len(entries) != 1 || entries[0].Path != "staged" {
		t.Errorf("focusedEntries() for staged = %v, want [{staged}]", entries)
	}

	m.focus = SectionUnstaged
	entries = m.focusedEntries()
	if len(entries) != 1 || entries[0].Path != "unstaged" {
		t.Errorf("focusedEntries() for unstaged = %v, want [{unstaged}]", entries)
	}

	m.focus = SectionGraph
	entries = m.focusedEntries()
	if entries != nil {
		t.Errorf("focusedEntries() for graph = %v, want nil", entries)
	}
}

func TestTruncateMsg(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer message", 15, "this is a lo..."},
		{"ab", 5, "ab"},
	}

	for _, tt := range tests {
		result := truncateMsg(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncateMsg(%q, %d) = %q, want %q", tt.input, tt.max, result, tt.expected)
		}
	}
}

func TestStopLoading(t *testing.T) {
	m, err := New(".")
	if err != nil {
		t.Fatalf("New(.) failed: %v", err)
	}

	m.loading = true
	m.loadingLabel = "Testing..."
	m.spinner.active = true

	m.stopLoading()

	if m.loading {
		t.Error("loading should be false")
	}
	if m.loadingLabel != "" {
		t.Errorf("loadingLabel = %q, want empty", m.loadingLabel)
	}
	if m.spinner.active {
		t.Error("spinner.active should be false")
	}
}

func TestSectionConstants(t *testing.T) {
	if SectionCommit != 0 {
		t.Errorf("SectionCommit = %d, want 0", SectionCommit)
	}
	if SectionStaged != 1 {
		t.Errorf("SectionStaged = %d, want 1", SectionStaged)
	}
	if SectionUnstaged != 2 {
		t.Errorf("SectionUnstaged = %d, want 2", SectionUnstaged)
	}
}

func TestModeConstants(t *testing.T) {
	if ModeNormal != 0 {
		t.Errorf("ModeNormal = %d, want 0", ModeNormal)
	}
	if ModeCommit != 1 {
		t.Errorf("ModeCommit = %d, want 1", ModeCommit)
	}
	if ModeDiff != 2 {
		t.Errorf("ModeDiff = %d, want 2", ModeDiff)
	}
	if ModeBranch != 3 {
		t.Errorf("ModeBranch = %d, want 3", ModeBranch)
	}
	if ModeHelp != 4 {
		t.Errorf("ModeHelp = %d, want 4", ModeHelp)
	}
	if ModeSettings != 5 {
		t.Errorf("ModeSettings = %d, want 5", ModeSettings)
	}
}
