package ui

import (
	"fmt"
	"strings"

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
	pointer      int
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
			if m.text != "" {
				var cmd tea.Cmd
				if m.onComplete != nil {
					onComplete := m.onComplete
					text := m.text
					cmd = func() tea.Msg {
						return onComplete(text)
					}
				}
				m.text = ""
				m.pointer = 0
				return m, cmd
			}
		} else if len(msgString) == 1 {
			m.text = m.text[:m.pointer] + msgString + m.text[m.pointer:]
			m.pointer++
		} else if msg.Type == tea.KeyRunes {
			toAdd := strings.Trim(msgString, "[]")
			m.text = m.text[:m.pointer] + toAdd + m.text[m.pointer:]
			m.pointer += len(toAdd)
		} else if msgString == "space" {
			m.text = m.text[:m.pointer] + " " + m.text[m.pointer:]
			m.pointer++
		} else if msgString == "backspace" {
			if m.pointer > 0 {
				m.text = m.text[:m.pointer-1] + m.text[m.pointer:]
				m.pointer = m.pointer - 1
			}
		} else if msgString == "left" {
			m.pointer = max(0, m.pointer-1)
		} else if msgString == "right" {
			m.pointer = min(len(m.text), m.pointer+1)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m textBox) View() string {
	style := lipgloss.NewStyle().MaxWidth(m.width).Width(m.width)
	arrowStyle := lipgloss.NewStyle()
	textStyle := lipgloss.NewStyle()
	finalCharStyle := lipgloss.NewStyle()
	var promptText string
	promptTextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	var text string
	if m.enabled {
		text = m.text
		arrowStyle = arrowStyle.Foreground(lipgloss.Color("5"))
		finalCharStyle = finalCharStyle.Background(lipgloss.Color("7"))
		if m.text == "" {
			promptText = "Talk to me..."
		}
	} else {
		text = m.disabledText
		textStyle = textStyle.Foreground(lipgloss.Color("8"))
		arrowStyle = arrowStyle.Foreground(lipgloss.Color("8"))
		finalCharStyle = finalCharStyle.Background(lipgloss.Color("8"))
	}
	text += " "
	textBeforePointer := text[:m.pointer]
	textOfPointer := text[m.pointer : m.pointer+1]
	textAfterPointer := text[m.pointer+1:]
	text = textStyle.Render(textBeforePointer) + finalCharStyle.Render(textOfPointer) + textStyle.Render(textAfterPointer)

	arrow := arrowStyle.Render("❯")
	promptText = promptTextStyle.Render(promptText)

	return style.Render(fmt.Sprintf("%s %s%s", arrow, text, promptText))
}
