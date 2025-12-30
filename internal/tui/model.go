package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/esferadigital/cadence/internal/config"
	"github.com/esferadigital/cadence/internal/logs"
	"github.com/esferadigital/cadence/internal/pomodoro"
	"github.com/esferadigital/cadence/internal/tui/navigation"
	"github.com/esferadigital/cadence/internal/tui/views/configview"
	"github.com/esferadigital/cadence/internal/tui/views/defaultview"
)

type model struct {
	logger logs.Logger
	width  int
	height int
	nav    navigation.Navigator
}

func newModel(machine *pomodoro.Machine, cfg config.Config, appLogger logs.Logger) model {
	return model{
		logger: appLogger,
		nav: navigation.New(
			navigation.ViewID("default"),
			map[navigation.ViewID]tea.Model{
				navigation.ViewID("default"): defaultview.New(machine),
				navigation.ViewID("config"):  configview.New(cfg),
			}),
	}
}

func (m model) Init() tea.Cmd {
	if current := m.nav.Current(); current != nil {
		return current.Init()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.logger != nil {
		m.logger.Printf("tui update: %T %+v", msg, msg)
	}

	switch msg := msg.(type) {
	case navigation.Msg:
		cmd, _ := m.nav.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	currentID := m.nav.CurrentID()
	if current := m.nav.Current(); current != nil {
		updated, cmd := current.Update(msg)
		// Persist the updated submodel in the navigator map; Bubble Tea returns a new model on Update.
		m.nav.SetView(currentID, updated)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	content := ""
	if current := m.nav.Current(); current != nil {
		content = current.View()
	}
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}
