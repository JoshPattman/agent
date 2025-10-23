package ui

import (
	"fmt"
	"strings"

	"github.com/JoshPattman/jpf"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func NewSummary(sum AgentSummary) tea.Model {
	return summary{
		summary: sum,
	}
}

type summary struct {
	summary AgentSummary
	width   int
	height  int
	usage   jpf.Usage
}

func (summary) Init() tea.Cmd {
	return nil
}

func (m summary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetWidth:
		m.width = msg.Width
		return m, nil
	case SetHeight:
		m.height = msg.Height
		return m, nil
	case UsageMessage:
		m.usage = msg.Usage
		return m, nil
	default:
		return m, nil
	}
}

func (m summary) View() string {
	boxStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("8")).
		Padding(1)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)
	modelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Bold(true)
	topRowContent := headerStyle.Render(m.summary.Name) + " ~ " + modelStyle.Render(m.summary.ModelName)
	topRow := lipgloss.NewStyle().Width(m.width-2).Border(lipgloss.ASCIIBorder(), false, false, true, false).AlignHorizontal(lipgloss.Center).Render(topRowContent)
	content := fmt.Sprintf(
		"%s\n%s\n%d MCP tools and %d subagents.\n\nInput: %d\nOutput: %d",
		topRow,
		strings.Join(m.summary.Description, " "),
		m.summary.NumMCP, m.summary.NumSubAgents,
		m.usage.InputTokens,
		m.usage.OutputTokens,
	)
	content = boxStyle.Render(content)
	return content
}
