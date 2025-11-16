package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewChat() tea.Model {
	return chat{
		height: 10,
		width:  10,
	}
}

type chat struct {
	messages     []tea.Model
	height       int
	width        int
	scrollOffset int
}

func (m chat) Init() tea.Cmd {
	return nil
}

func (m chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ResetScrollMessage:
		m.scrollOffset = 0
		return m, nil
	case ResetMessages:
		m.messages = make([]tea.Model, 0)
		return m, nil
	case ScrollMessage:
		m.scrollOffset += msg.Delta
		m.scrollOffset = max(m.scrollOffset, 0)
		return m, nil
	case SetWidth:
		for i := range m.messages {
			newMessage, _ := m.messages[i].Update(msg)
			m.messages[i] = newMessage
		}
		m.width = msg.Width
		return m, nil
	case SetHeight:
		m.height = msg.Height
		return m, nil
	case AddMessage:
		newMessage := Message{
			msg.Type,
			msg.Content,
			m.width,
		}
		m.messages = append(m.messages, newMessage)
		return m, nil
	default:
		return m, nil
	}
}

func (m chat) View() string {
	renderedMessages := make([]string, len(m.messages))
	for i, msg := range m.messages {
		renderedMessages[i] = msg.View()
	}
	fullContent := lipgloss.JoinVertical(lipgloss.Left, renderedMessages...)
	fullContent = truncateHeightWithOffset(fullContent, m.height, m.scrollOffset)
	style := lipgloss.NewStyle().Height(m.height).AlignVertical(lipgloss.Bottom)
	return style.Render(fullContent)
}

func truncateHeightWithOffset(s string, height, offset int) string {
	lines := strings.Split(s, "\n")

	if offset >= len(lines) {
		return ""
	}

	offset = max(0, offset)
	end := len(lines) - offset
	start := end - height
	start = max(start, 0)

	truncated := lines[start:end]
	return strings.Join(truncated, "\n")
}
