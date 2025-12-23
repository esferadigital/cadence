package tui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/esferadigital/cadence/internal/logs"
	"github.com/esferadigital/cadence/internal/pomodoro"
)

type model struct {
	phase      pomodoro.PhaseSnapshot
	workPhases int
	done       bool
	status     pomodoro.TimerStatus
	machine    *pomodoro.Machine
	logger     logs.Logger
	width      int
	height     int
	blinkOn    bool
}

var (
	indicatorWidth  = 20
	indicatorHeight = 1
)

const (
	indicatorOn  = "█"
	indicatorOff = "░"
)

func newModel(machine *pomodoro.Machine, appLogger logs.Logger) model {
	return model{machine: machine, logger: appLogger}
}

func (m model) Init() tea.Cmd {
	return m.getStateCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.logger != nil {
		m.logger.Printf("tui update: %T %+v", msg, msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "s":
			return m, func() tea.Msg {
				m.machine.Start()
				return nil
			}
		case "p":
			return m, func() tea.Msg {
				m.machine.Pause()
				return nil
			}
		case "r":
			return m, func() tea.Msg {
				m.machine.Resume()
				return nil
			}
		}
	case pomodoro.EventStateChanged:
		phaseChanged := msg.Phase.Idx != m.phase.Idx || msg.Phase.Kind != m.phase.Kind
		m.phase = msg.Phase
		m.status = msg.Status
		m.workPhases = msg.WorkPhases
		if m.status == pomodoro.StatusRunning {
			if phaseChanged {
				m.blinkOn = true
			} else {
				m.blinkOn = !m.blinkOn
			}
		} else {
			m.blinkOn = true
		}
		return m, nil
	case pomodoro.EventTimerFinished:
		m.done = true
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	var content string
	if m.done {
		content = "Nice job!\n\n[q] quit"
	} else {
		indicator := renderPhaseIndicator(m.phase, m.status, m.workPhases, m.blinkOn)
		content = fmt.Sprintf("%s\n\n%s\n\n%s", renderRemaining(m.phase.Remaining), indicator, m.hints())
	}
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m model) getStateCmd() tea.Cmd {
	return func() tea.Msg {
		m.machine.GetState()
		return nil
	}
}

func (m model) hints() string {
	hints := make([]string, 0, 2)
	switch m.status {
	case pomodoro.StatusInit:
		hints = append(hints, "[s] start")
	case pomodoro.StatusRunning:
		hints = append(hints, "[p] pause")
	case pomodoro.StatusPaused:
		hints = append(hints, "[r] resume")
	}
	hints = append(hints, "[q] quit")
	return strings.Join(hints, "  ")
}

func renderPhaseIndicator(phase pomodoro.PhaseSnapshot, status pomodoro.TimerStatus, workPhases int, blinkOn bool) string {
	// Breaks show a textual indicator instead of phase indicators.
	if phase.Kind == pomodoro.PhaseBreak {
		return indicatorBox(fmt.Sprintf("break %d", phase.HumanIdx))
	}

	// Work phases show indicators.
	// Handle a 0-value `workPhases`, which can happen before the first state update
	var indicatorCnt int
	if workPhases <= 0 {
		indicatorCnt = 1
	} else {
		indicatorCnt = workPhases
	}
	indicators := make([]string, indicatorCnt)

	// Initial state: all indicators off.
	if status == pomodoro.StatusInit {
		for i := 0; i < indicatorCnt; i++ {
			indicators[i] = indicatorOff
		}
		return indicatorBox(strings.Join(indicators, " "))
	}

	// Completed work phases stay on. Pending phases are off. Current phase is blinking.
	activeIdx := phase.HumanIdx - 1
	for i := 0; i < indicatorCnt; i++ {
		if i < activeIdx {
			indicators[i] = indicatorOn
		} else {
			indicators[i] = indicatorOff
		}
	}

	if activeIdx >= 0 && activeIdx < indicatorCnt {
		active := indicatorOn
		if status == pomodoro.StatusRunning && !blinkOn {
			active = indicatorOff
		}
		indicators[activeIdx] = active
	}

	return indicatorBox(strings.Join(indicators, " "))
}

func indicatorBox(content string) string {
	width := indicatorWidth
	height := indicatorHeight
	contentWidth := lipgloss.Width(content)
	if width < contentWidth {
		width = contentWidth
	}
	if height < 1 {
		height = 1
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func formatRemaining(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	if totalSeconds < 0 {
		totalSeconds = -totalSeconds
	}
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func renderRemaining(d time.Duration) string {
	timeText := formatRemaining(d)
	glyphs := make([][]string, 0, len(timeText))
	maxHeight := 0
	for _, r := range timeText {
		lines := glyphLinesForRune(r)
		if len(lines) > maxHeight {
			maxHeight = len(lines)
		}
		glyphs = append(glyphs, lines)
	}

	var b strings.Builder
	for row := 0; row < maxHeight; row++ {
		for idx, glyph := range glyphs {
			if row < len(glyph) {
				b.WriteString(glyph[row])
			}
			if idx < len(glyphs)-1 {
				b.WriteString(" ")
			}
		}
		if row < maxHeight-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func glyphLinesForRune(r rune) []string {
	glyph, ok := digitMap[string(r)]
	if !ok {
		return normalizeGlyphLines([]string{string(r)})
	}
	lines := append([]string(nil), glyph...)
	return normalizeGlyphLines(lines)
}

func normalizeGlyphLines(lines []string) []string {
	maxWidth := 0
	for _, line := range lines {
		if width := utf8.RuneCountInString(line); width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth == 0 {
		return lines
	}
	for i, line := range lines {
		if width := utf8.RuneCountInString(line); width < maxWidth {
			lines[i] = line + strings.Repeat(" ", maxWidth-width)
		}
	}
	return lines
}
