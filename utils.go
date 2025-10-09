package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// convertActionArgsToMap converts the new Action.Args structure to map[string]any for tool calls
func convertActionArgsToMap(args []ActionArg) map[string]any {
	out := make(map[string]any)

	for _, arg := range args {
		out[arg.ArgName] = arg.ArgData
	}

	return out
}

// FormatActionArgsForDisplay formats the new Action.Args structure for display purposes
func FormatActionArgsForDisplay(args []ActionArg) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%s=%v", arg.ArgName, arg.ArgData)
	}

	return strings.Join(parts, "&")
}

func formatTemplate(tpl string, data any) (string, error) {
	tplCompiled, err := template.New("prompt").Parse(tpl)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(nil)
	err = tplCompiled.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
