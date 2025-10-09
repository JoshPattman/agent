package agentmcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JoshPattman/agent"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Create an MCP HTTP client and initialise it
func CreateClient(addr string, customHeaders map[string]string) (*client.Client, error) {
	httpTransport, err := transport.NewStreamableHTTP(
		addr,
		transport.WithHTTPHeaders(customHeaders),
	)
	if err != nil {
		return nil, err
	}
	c := client.NewClient(httpTransport)

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Agent",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	_, err = c.Initialize(context.Background(), initRequest)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Get the tools from the MCP client and convert them to agent tools
func CreateToolsFromMCP(client *client.Client) ([]agent.Tool, error) {
	ctx := context.Background()
	result, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	tools := make([]agent.Tool, len(result.Tools))
	for i, mcpTool := range result.Tools {
		agentTool, err := createTool(client, mcpTool)
		if err != nil {
			return nil, err
		}
		tools[i] = agentTool
	}
	return tools, nil
}

func createTool(client *client.Client, tool mcp.Tool) (agent.Tool, error) {
	return &mcpTool{client, tool}, nil
}

type mcpTool struct {
	client *client.Client
	tool   mcp.Tool
}

// Call implements agent.Tool.
func (m *mcpTool) Call(args map[string]any) (string, error) {
	for key, arg := range args {
		_, ok := arg.([]any)
		if ok {
			continue
		}
		argProp, ok := m.tool.InputSchema.Properties[key]
		if !ok {
			return "", fmt.Errorf("argument %s not wanted", key)
		}
		argPropDict, ok := argProp.(map[string]any)
		if !ok {
			continue // Weird format but we will survive
		}
		if argPropDict["type"] == "array" {
			fmt.Println("converted", key, "to array")
			args[key] = arg
		}
	}
	res, err := m.client.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      m.tool.Name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	contents := make([]string, 0)
	for _, c := range res.Content {
		if c, ok := c.(mcp.TextContent); ok {
			contents = append(contents, c.Text)
		}
	}
	if len(contents) == 0 {
		return "", errors.New("tool returned no content")
	}
	return strings.Join(contents, "\n\n\n"), nil
}

// Description implements agent.Tool.
func (m *mcpTool) Description() []string {
	desc := []string{m.tool.Description}
	for pname, prop := range m.tool.InputSchema.Properties {
		desc = append(
			desc,
			fmt.Sprintf("Param `%s` (%s): %s", pname, prop.(map[string]any)["type"].(string), prop.(map[string]any)["description"].(string)),
		)
	}
	return desc
}

// Name implements agent.Tool.
func (m *mcpTool) Name() string {
	return m.tool.Name
}
