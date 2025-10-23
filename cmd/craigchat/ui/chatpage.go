package ui

import (
	"fmt"
	"strings"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/agent/craig"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AgentSummary struct {
	Name         string
	Description  []string
	NumMCP       int
	NumSubAgents int
	ModelName    string
}

func NewChatPage(buildAgent func() (agent.Agent, error), summary AgentSummary) tea.Model {
	cp := chatPage{
		10,
		10,
		NewChat(),
		NewTextBox(),
		buildAgent,
		nil,
		nil,
		NewSummary(summary),
	}
	cp.textInput, _ = cp.textInput.Update(SetTextboxCompleteMessage{
		func(s string) tea.Msg {
			return UserMessageSend{s}
		},
	})
	return cp
}

type chatPage struct {
	width       int
	height      int
	chat        tea.Model
	textInput   tea.Model
	buildAgent  func() (agent.Agent, error)
	activeAgent agent.Agent
	sendConcMsg func(tea.Msg)
	summary     tea.Model
}

func (chatPage) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ResetAgentMessage{}
		},
	)
}

func (m chatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		summaryWidth := 50
		mainWidth := m.width - summaryWidth
		if mainWidth < 5 {
			mainWidth = 5
		}
		m.chat, _ = m.chat.Update(SetWidth{mainWidth - 2})
		m.chat, _ = m.chat.Update(SetHeight{m.height - 3})
		m.textInput, _ = m.textInput.Update(SetWidth{mainWidth - 2})
		m.summary, _ = m.summary.Update(SetWidth{summaryWidth - 3})
		m.summary, _ = m.summary.Update(SetHeight{m.height - 2})
		return m, nil
	case UserMessageSend:
		m.chat, _ = m.chat.Update(AddMessage{UserMessage, msg.Message})
		m.textInput, _ = m.textInput.Update(EnableMessage{false})
		activeAgent := m.activeAgent
		cmd := func() tea.Msg {
			result, err := activeAgent.Answer(msg.Message)
			if err != nil {
				panic(err) // TODO: Handle gracefully
			}
			return AIMessageSend{result}
		}
		return m, cmd
	case AIMessageSend:
		m.chat, _ = m.chat.Update(AddMessage{CRAIGMessage, msg.Message})
		m.textInput, _ = m.textInput.Update(EnableMessage{true})
		return m, nil
	case AIReasoningSend:
		var text string
		if len(msg.ToolCalls) > 0 {
			text = fmt.Sprintf("%s -> %s", msg.Reasoning, strings.Join(msg.ToolCalls, ", "))
		} else {
			text = msg.Reasoning
		}
		m.chat, _ = m.chat.Update(AddMessage{CRAIGReasoningMessage, text})
		return m, nil
	case ResetAgentMessage:
		m.chat, _ = m.chat.Update(ResetMessages{})
		newAgent, err := m.buildAgent()
		if err != nil {
			panic(err) // TODO: Deal with this gracefully
		}
		sendConcMsg := m.sendConcMsg
		newAgent.SetOnReActCompleteCallback(func(s string, ao []agent.ActionObservation) {
			toolCalls := make([]string, len(ao))
			for i := range ao {
				aa := make([]agent.ActionArg, len(ao[i].Action.Args))
				for j := range aa {
					aa[j] = ao[i].Action.Args[j]
				}
				toolCalls[i] = fmt.Sprintf("%s?%s", ao[i].Action.Name, craig.FormatActionArgsForDisplay(aa))
			}
			if sendConcMsg != nil {
				sendConcMsg(AIReasoningSend{s, toolCalls})
			}
		})
		m.activeAgent = newAgent
		return m, nil
	case SetConcurrentMessageSender:
		m.sendConcMsg = msg.MsgSender
		return m, nil
	case UsageMessage:
		m.summary, _ = m.summary.Update(msg)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			m.chat, _ = m.chat.Update(ScrollMessage{1})
			return m, nil
		case "down":
			m.chat, _ = m.chat.Update(ScrollMessage{-1})
			return m, nil
		case "esc":
			m.chat, _ = m.chat.Update(ResetMessages{})
			return m, nil
		default:
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
	default:
		return m, nil
	}
}

func (m chatPage) View() string {
	mainStyle := lipgloss.NewStyle().
		Width(m.width - 2).
		Height(m.height - 2).
		BorderForeground(lipgloss.Color("8")).
		Border(lipgloss.DoubleBorder())
	content := lipgloss.JoinVertical(lipgloss.Left, m.chat.View(), m.textInput.View())
	content = lipgloss.JoinHorizontal(lipgloss.Bottom, content, m.summary.View())
	return mainStyle.Render(content)
}
