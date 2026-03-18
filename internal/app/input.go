package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInput is a minimal single-line text input with cursor.
type TextInput struct {
	Value       string
	Cursor      int
	Focused     bool
	Disabled    bool
	Placeholder string
}

func NewTextInput(placeholder string) TextInput {
	return TextInput{Placeholder: placeholder}
}

func (t *TextInput) Focus() {
	t.Focused = true
}

func (t *TextInput) Blur() {
	t.Focused = false
}

func (t *TextInput) Reset() {
	t.Value = ""
	t.Cursor = 0
}

func (t *TextInput) SetValue(s string) {
	t.Value = s
	t.Cursor = len([]rune(s))
}

func (t *TextInput) Update(msg tea.KeyMsg) {
	if t.Disabled {
		return
	}
	runes := []rune(t.Value)
	key := msg.String()

	switch key {
	case "left":
		if t.Cursor > 0 {
			t.Cursor--
		}
	case "right":
		if t.Cursor < len(runes) {
			t.Cursor++
		}
	case "home", "ctrl+a":
		t.Cursor = 0
	case "end", "ctrl+e":
		t.Cursor = len(runes)
	case "backspace":
		if t.Cursor > 0 {
			runes = append(runes[:t.Cursor-1], runes[t.Cursor:]...)
			t.Value = string(runes)
			t.Cursor--
		}
	case "delete", "ctrl+d":
		if t.Cursor < len(runes) {
			runes = append(runes[:t.Cursor], runes[t.Cursor+1:]...)
			t.Value = string(runes)
		}
	case "ctrl+k":
		// Kill to end of line
		t.Value = string(runes[:t.Cursor])
	case "ctrl+u":
		// Kill to start of line
		t.Value = string(runes[t.Cursor:])
		t.Cursor = 0
	case "ctrl+w":
		// Delete word backward
		if t.Cursor > 0 {
			i := t.Cursor - 1
			for i > 0 && runes[i-1] == ' ' {
				i--
			}
			for i > 0 && runes[i-1] != ' ' {
				i--
			}
			t.Value = string(append(runes[:i], runes[t.Cursor:]...))
			t.Cursor = i
		}
	default:
		// Type character if it's a single rune
		if len(msg.Runes) > 0 && !strings.HasPrefix(key, "ctrl+") && !strings.HasPrefix(key, "alt+") {
			for _, r := range msg.Runes {
				runes = append(runes[:t.Cursor], append([]rune{r}, runes[t.Cursor:]...)...)
				t.Cursor++
			}
			t.Value = string(runes)
		}
	}
}

// View renders the text input content (without border — caller wraps it).
func (t TextInput) View(width int) string {
	if !t.Focused && t.Value == "" {
		return t.Placeholder
	}

	runes := []rune(t.Value)

	// Scroll if cursor is past the visible area
	visibleStart := 0
	if t.Cursor > width-2 {
		visibleStart = t.Cursor - width + 2
	}
	visibleEnd := visibleStart + width - 1
	if visibleEnd > len(runes) {
		visibleEnd = len(runes)
	}

	visible := runes[visibleStart:visibleEnd]
	cursorPos := t.Cursor - visibleStart

	if !t.Focused || t.Disabled {
		return string(visible)
	}

	// Render with cursor block
	var out strings.Builder
	for i, r := range visible {
		if i == cursorPos {
			out.WriteString("\033[7m") // reverse video (block cursor)
			out.WriteRune(r)
			out.WriteString("\033[0m")
		} else {
			out.WriteRune(r)
		}
	}
	// If cursor is at the end, show cursor on a space
	if cursorPos >= len(visible) {
		out.WriteString("\033[7m \033[0m")
	}

	return out.String()
}
