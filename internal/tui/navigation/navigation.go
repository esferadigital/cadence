package navigation

import tea "github.com/charmbracelet/bubbletea"

type ViewID string

type Action int

const (
	ActionPush Action = iota
	ActionPop
	ActionReplace
	ActionReset
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

func Replace(view ViewID) Msg {
	return Msg{Action: ActionReplace, View: view}
}

func Reset(view ViewID) Msg {
	return Msg{Action: ActionReset, View: view}
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

func ReplaceCmd(view ViewID) tea.Cmd {
	return func() tea.Msg {
		return Replace(view)
	}
}

func ResetCmd(view ViewID) tea.Cmd {
	return func() tea.Msg {
		return Reset(view)
	}
}

type Navigator struct {
	defaultID ViewID
	stack     []ViewID
	views     map[ViewID]tea.Model
}

func New(defaultID ViewID, views map[ViewID]tea.Model) Navigator {
	return Navigator{
		defaultID: defaultID,
		stack:     []ViewID{defaultID},
		views:     views,
	}
}

func (n *Navigator) CurrentID() ViewID {
	if len(n.stack) == 0 {
		return n.defaultID
	}
	return n.stack[len(n.stack)-1]
}

func (n *Navigator) Current() tea.Model {
	if current, ok := n.views[n.CurrentID()]; ok {
		return current
	}
	if fallback, ok := n.views[n.defaultID]; ok {
		return fallback
	}
	return nil
}

func (n *Navigator) SetView(id ViewID, model tea.Model) {
	if n.views == nil {
		n.views = make(map[ViewID]tea.Model)
	}
	n.views[id] = model
}

func (n *Navigator) Update(msg tea.Msg) (tea.Cmd, bool) {
	navMsg, ok := msg.(Msg)
	if !ok {
		return nil, false
	}

	switch navMsg.Action {
	case ActionPush:
		if _, ok := n.views[navMsg.View]; !ok {
			return nil, true
		}
		n.stack = append(n.stack, navMsg.View)
		return n.views[navMsg.View].Init(), true
	case ActionPop:
		if len(n.stack) <= 1 {
			n.stack = []ViewID{n.defaultID}
			return nil, true
		}
		n.stack = n.stack[:len(n.stack)-1]
		return nil, true
	case ActionReplace:
		if _, ok := n.views[navMsg.View]; !ok {
			return nil, true
		}
		if len(n.stack) == 0 {
			n.stack = []ViewID{navMsg.View}
		} else {
			n.stack[len(n.stack)-1] = navMsg.View
		}
		return n.views[navMsg.View].Init(), true
	case ActionReset:
		if _, ok := n.views[navMsg.View]; !ok {
			return nil, true
		}
		n.stack = []ViewID{navMsg.View}
		return n.views[navMsg.View].Init(), true
	default:
		return nil, true
	}
}
