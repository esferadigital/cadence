package configview

import (
	"errors"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

func New() *Model {
	m := &Model{}
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
		// TODO: apply config changes on save.
		return m, navigation.PopCmd()
	}

	return m, cmd
}

func (m *Model) View() string {
	return m.form.View()
}

func (m *Model) initConfigForm() {
	if m.config.workMinutes == "" {
		m.config = configState{
			workMinutes:  "25",
			phaseMinutes: "5",
			workPhases:   "4",
		}
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
