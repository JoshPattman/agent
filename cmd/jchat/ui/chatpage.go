package ui

import (
	"fmt"
	"strings"
	"time"

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
	t := time.Now()
	cp := chatPage{
		10,
		10,
		NewChat(),
		NewTextBox(),
		buildAgent,
		nil,
		&t,
		nil,
		NewSummary(summary),
		false,
	}
	cp.textInput, _ = cp.textInput.Update(SetTextboxCompleteMessage{
		func(s string) tea.Msg {
			return UserMessageSend{s}
		},
	})
	return cp
}

type chatPage struct {
	width               int
	height              int
	chat                tea.Model
	textInput           tea.Model
	buildAgent          func() (agent.Agent, error)
	activeAgent         agent.Agent
	lastUserMessageTime *time.Time
	sendConcMsg         func(tea.Msg)
	summary             tea.Model
	awaitingResponse    bool
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
	case SetThinkingNumDots:
		if !m.awaitingResponse {
			return m, nil
		} else {
			m.chat, _ = m.chat.Update(SetChatInfoMessage{
				fmt.Sprintf("Thinking%s", strings.Repeat(".", msg.N)),
			})
			return m, func() tea.Msg {
				time.Sleep(time.Second / 4)
				return SetThinkingNumDots{(msg.N + 1) % 4}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		summaryWidth := 40
		mainWidth := max(m.width-summaryWidth, 5)
		m.chat, _ = m.chat.Update(SetWidth{mainWidth})
		m.chat, _ = m.chat.Update(SetHeight{m.height})
		m.textInput, _ = m.textInput.Update(SetWidth{mainWidth})
		m.summary, _ = m.summary.Update(SetWidth{summaryWidth})
		m.summary, _ = m.summary.Update(SetHeight{m.height})
		return m, nil
	case UserMessageSend:
		m.chat, _ = m.chat.Update(AddMessage{UserMessage, msg.Message})
		m.textInput, _ = m.textInput.Update(EnableMessage{false})
		*m.lastUserMessageTime = time.Now()
		m.awaitingResponse = true
		activeAgent := m.activeAgent
		cmd := func() tea.Msg {
			result, err := activeAgent.Answer(msg.Message)
			if err != nil {
				return AIErrorSend{err}
			}
			return AIMessageSend{result}
		}
		cmdDots := func() tea.Msg {
			return SetThinkingNumDots{1}
		}
		m.chat, _ = m.chat.Update(SetChatInfoMessage{"Thinking..."})
		return m, tea.Batch(cmd, cmdDots)
	case AIMessageSend:
		m.chat, _ = m.chat.Update(AddMessage{CRAIGMessage, msg.Message})
		m.textInput, _ = m.textInput.Update(EnableMessage{true})
		return m, nil
	case AIReasoningSend:
		var text string
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]string, len(msg.ToolCalls))
			for i := range msg.ToolCalls {
				toolCalls[i] = fmt.Sprintf("  └▶ %s", msg.ToolCalls[i])
			}
			text = fmt.Sprintf("%s\n%s", "Called tools", strings.Join(toolCalls, "\n"))
		} else {
			text = fmt.Sprintf("Thougt for %s", formatDuration1dp(msg.For))
		}
		m.chat, _ = m.chat.Update(AddMessage{CRAIGReasoningMessage, text})
		m.awaitingResponse = false
		m.chat, _ = m.chat.Update(SetChatInfoMessage{""})
		return m, nil
	case AIErrorSend:
		m.chat, _ = m.chat.Update(AddMessage{ErrorMessage, msg.Error.Error()})
		m.textInput, _ = m.textInput.Update(EnableMessage{true})
		m.awaitingResponse = false
		m.chat, _ = m.chat.Update(SetChatInfoMessage{""})
		return m, nil
	case ResetAgentMessage:
		m.chat, _ = m.chat.Update(ResetMessages{})
		newAgent, err := m.buildAgent()
		if err != nil {
			cmd := func() tea.Msg { return AIErrorSend{err} }
			return m, cmd
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
				sendConcMsg(AIReasoningSend{s, toolCalls, time.Since(*m.lastUserMessageTime)})
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
		Width(m.width).
		Height(m.height)
	content := lipgloss.JoinVertical(lipgloss.Left, m.chat.View(), m.textInput.View())
	content = lipgloss.JoinHorizontal(lipgloss.Bottom, content, m.summary.View())
	return mainStyle.Render(content)
}

func formatDuration1dp(d time.Duration) string {
	secs := float64(d) / float64(time.Second)
	return fmt.Sprintf("%.1fs", secs)
}
