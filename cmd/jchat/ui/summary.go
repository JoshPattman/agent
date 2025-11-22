package ui

import (
	"fmt"
	"strings"

	"github.com/JoshPattman/jpf"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
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
	header := figure.NewFigure("JChat", "ogre", true).String()
	header = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Width(m.width - 1).
		AlignHorizontal(lipgloss.Center).
		Render(strings.TrimSpace(header))

	boxStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1)

	agentName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		Render(m.summary.Name)
	modelName := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true).
		Render(m.summary.ModelName)
	topRow := lipgloss.NewStyle().
		Width(m.width - 1).
		AlignHorizontal(lipgloss.Center).
		Render(agentName + " ~ " + modelName)
	ioText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("I/O")
	ioBlock := lipgloss.JoinVertical(
		lipgloss.Center,
		fmt.Sprintf("%d/%d", m.usage.InputTokens, m.usage.OutputTokens),
		ioText,
	)
	ioBlock = lipgloss.NewStyle().Width(m.width - 1).AlignHorizontal(lipgloss.Center).Render(ioBlock)
	content := fmt.Sprintf(
		"%s\n\n%s\n%s\n%d MCP servers and %d subagents.\n\n%s",
		header,
		topRow,
		strings.Join(m.summary.Description, " "),
		m.summary.NumMCP, m.summary.NumSubAgents,
		ioBlock,
	)
	content = boxStyle.Render(content)
	return content
}
