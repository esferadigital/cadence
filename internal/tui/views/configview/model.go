package configview

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/esferadigital/cadence/internal/config"
	"github.com/esferadigital/cadence/internal/tui/navigation"
)

type Model struct {
	config configState
	form   *huh.Form
}

type configState struct {
	workMinutes  string
	phaseMinutes string
	workPhases   string
}

func New(cfg config.Config) *Model {
	m := &Model{
		config: configStateFromConfig(cfg),
	}
	m.initConfigForm()
	return m
}

func (m *Model) Init() tea.Cmd {
	m.initConfigForm()
	return m.form.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, navigation.PopCmd()
		case "q":
			return m, tea.Quit
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		cfg, err := configFromState(m.config)
		if err != nil {
			return m, tea.Batch(tea.Printf("invalid config: %v\n", err), tea.Quit)
		}
		if err := config.Save(cfg); err != nil {
			return m, tea.Batch(tea.Printf("failed to save config: %v\n", err), tea.Quit)
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *Model) View() string {
	return m.form.View() + "\n\n[esc] close config  [q] quit  Saving closes the program."
}

func (m *Model) initConfigForm() {
	if m.config.workMinutes == "" {
		m.config = configStateFromConfig(config.Default())
	}
	m.form = newConfigForm(&m.config)
}

func newConfigForm(config *configState) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Work minutes").
				Value(&config.workMinutes).
				Validate(validatePositiveInt),
			huh.NewInput().
				Title("Phase minutes").
				Value(&config.phaseMinutes).
				Validate(validatePositiveInt),
			huh.NewInput().
				Title("Work phases").
				Value(&config.workPhases).
				Validate(validatePositiveInt),
		),
	)
}

func validatePositiveInt(value string) error {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return errors.New("enter a positive whole number")
	}
	return nil
}

func configStateFromConfig(cfg config.Config) configState {
	return configState{
		workMinutes:  strconv.Itoa(cfg.WorkMinutes),
		phaseMinutes: strconv.Itoa(cfg.BreakMinutes),
		workPhases:   strconv.Itoa(cfg.WorkPhases),
	}
}

func configFromState(state configState) (config.Config, error) {
	workMinutes, err := strconv.Atoi(strings.TrimSpace(state.workMinutes))
	if err != nil {
		return config.Config{}, fmt.Errorf("work minutes: %w", err)
	}
	breakMinutes, err := strconv.Atoi(strings.TrimSpace(state.phaseMinutes))
	if err != nil {
		return config.Config{}, fmt.Errorf("break minutes: %w", err)
	}
	workPhases, err := strconv.Atoi(strings.TrimSpace(state.workPhases))
	if err != nil {
		return config.Config{}, fmt.Errorf("work phases: %w", err)
	}

	return config.Config{
		WorkMinutes:  workMinutes,
		BreakMinutes: breakMinutes,
		WorkPhases:   workPhases,
	}, nil
}
