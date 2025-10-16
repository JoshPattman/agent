package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewTextBox() tea.Model {
	return textBox{enabled: true}
}

type textBox struct {
	enabled      bool
	text         string
	disabledText string
	width        int
	onComplete   func(string) tea.Msg
}

func (textBox) Init() tea.Cmd {
	return nil
}

func (m textBox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EnableMessage:
		m.enabled = msg.Enable
		return m, nil
	case SetWidth:
		m.width = msg.Width
		return m, nil
	case SetTextboxCompleteMessage:
		m.onComplete = msg.BuildOnComplete
		return m, nil
	case tea.KeyMsg:
		if !m.enabled {
			return m, nil
		}
		msgString := msg.String()
		if msgString == "enter" {
			var cmd tea.Cmd
			if m.onComplete != nil {
				onComplete := m.onComplete
				text := m.text
				cmd = func() tea.Msg {
					return onComplete(text)
				}
			}
			m.text = ""
			return m, cmd
		} else if len(msgString) == 1 {
			m.text += msgString
		} else if msgString == "space" {
			m.text += " "
		} else if msgString == "backspace" {
			if len(m.text) > 0 {
				m.text = m.text[:len(m.text)-1]
			}
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m textBox) View() string {
	style := lipgloss.NewStyle().MaxWidth(m.width).Width(m.width)

	var text string
	if m.enabled {
		text = m.text
	} else {
		text = m.disabledText
		style = style.Foreground(lipgloss.Color("8"))
	}

	return style.Render(fmt.Sprintf("> %s", text))
}
