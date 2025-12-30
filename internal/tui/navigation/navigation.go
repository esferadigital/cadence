package navigation

import tea "github.com/charmbracelet/bubbletea"

type ViewID string

const (
	ViewDefault ViewID = "default"
	ViewConfig  ViewID = "config"
)

type Action int

const (
	ActionPush Action = iota
	ActionPop
)

type Msg struct {
	Action Action
	View   ViewID
}

func Push(view ViewID) Msg {
	return Msg{Action: ActionPush, View: view}
}

func Pop() Msg {
	return Msg{Action: ActionPop}
}

func PushCmd(view ViewID) tea.Cmd {
	return func() tea.Msg {
		return Push(view)
	}
}

func PopCmd() tea.Cmd {
	return func() tea.Msg {
		return Pop()
	}
}
