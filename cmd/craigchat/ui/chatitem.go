package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MessageType int

const (
	UserMessage MessageType = iota
	CRAIGMessage
	CRAIGReasoningMessage
)

type Message struct {
	msgType MessageType
	content string
	width   int
}

func (m Message) Init() tea.Cmd {
	return nil
}

func (m Message) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetWidth:
		m.width = msg.Width
		return m, nil
	default:
		return m, nil
	}
}

func (c Message) View() string {
	var borderColor lipgloss.Color
	var textColor lipgloss.Color

	switch c.msgType {
	case UserMessage:
		borderColor = lipgloss.Color("4")
		textColor = lipgloss.Color("")
	case CRAIGMessage:
		borderColor = lipgloss.Color("2")
		textColor = lipgloss.Color("")
	case CRAIGReasoningMessage:
		borderColor = lipgloss.Color("3")
		textColor = lipgloss.Color("240")
	}

	style := lipgloss.NewStyle().
		Width(c.width-1).
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		BorderForeground(borderColor).
		Foreground(textColor).
		PaddingLeft(1).
		PaddingRight(1)
	content := style.Render(c.content)
	if c.msgType == UserMessage {
		content = "\n" + content
	}
	return content
}
