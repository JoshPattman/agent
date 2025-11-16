package ui

import (
	"github.com/JoshPattman/jpf"
	tea "github.com/charmbracelet/bubbletea"
)

type ScrollMessage struct {
	Delta int
}

type ResetScrollMessage struct{}

type SetWidth struct {
	Width int
}

type SetHeight struct {
	Height int
}

type AddMessage struct {
	Type    MessageType
	Content string
}

type ResetMessages struct{}

type EnableMessage struct {
	Enable bool
}

type SetTextboxCompleteMessage struct {
	BuildOnComplete func(string) tea.Msg
}

type UserMessageSend struct {
	Message string
}

type AIMessageSend struct {
	Message string
}

type AIReasoningSend struct {
	Reasoning string
	ToolCalls []string
}

type ResetAgentMessage struct{}

type SetConcurrentMessageSender struct {
	MsgSender func(tea.Msg)
}

type UsageMessage struct {
	Usage jpf.Usage
}
