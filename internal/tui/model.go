package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/esferadigital/cadence/internal/logs"
	"github.com/esferadigital/cadence/internal/pomodoro"
	"github.com/esferadigital/cadence/internal/tui/navigation"
	"github.com/esferadigital/cadence/internal/tui/views/configview"
	"github.com/esferadigital/cadence/internal/tui/views/defaultview"
)

type model struct {
	logger    logs.Logger
	width     int
	height    int
	viewStack []navigation.ViewID
	// Keep submodels as pointers so state persists across updates and view changes.
	viewMap map[navigation.ViewID]tea.Model
}

func newModel(machine *pomodoro.Machine, appLogger logs.Logger) model {
	m := model{
		logger:    appLogger,
		viewStack: []navigation.ViewID{navigation.ViewDefault},
		viewMap: map[navigation.ViewID]tea.Model{
			navigation.ViewDefault: defaultview.New(machine),
			navigation.ViewConfig:  configview.New(),
		},
	}
	return m
}

func (m model) Init() tea.Cmd {
	if current := m.currentView(); current != nil {
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
		return m.handleNavigation(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	currentID := m.currentViewID()
	if current, ok := m.viewMap[currentID]; ok {
		updated, cmd := current.Update(msg)
		m.viewMap[currentID] = updated
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	content := ""
	if current := m.currentView(); current != nil {
		content = current.View()
	}
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func (m model) currentViewID() navigation.ViewID {
	if len(m.viewStack) == 0 {
		return navigation.ViewDefault
	}
	return m.viewStack[len(m.viewStack)-1]
}

func (m model) currentView() tea.Model {
	if current, ok := m.viewMap[m.currentViewID()]; ok {
		return current
	}
	if fallback, ok := m.viewMap[navigation.ViewDefault]; ok {
		return fallback
	}
	return nil
}

func (m *model) pushView(view navigation.ViewID) {
	m.viewStack = append(m.viewStack, view)
}

func (m *model) popView() {
	if len(m.viewStack) <= 1 {
		m.viewStack = []navigation.ViewID{navigation.ViewDefault}
		return
	}
	m.viewStack = m.viewStack[:len(m.viewStack)-1]
}

func (m model) handleNavigation(msg navigation.Msg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case navigation.ActionPush:
		if _, ok := m.viewMap[msg.View]; !ok {
			return m, nil
		}
		m.pushView(msg.View)
		return m, m.viewMap[msg.View].Init()
	case navigation.ActionPop:
		m.popView()
		return m, nil
	default:
		return m, nil
	}
}
