# jgent

A command line tool for interacting with AI agents configured via JSON files.
- Attach MCP tools (text-tools only)
- Create sub agents recursively

## Usage

```bash
jgent [flags]
```

### Flags

- `-agent string`: Path to the agent config file (default: "./agent.json")
- `-models string`: Path to the models config file (default: "./models.json")

## Configuration Files

### agent.json

Configure the agent's model and MCP servers:

```json
{
    "agent_name": "josh's agent",
    "agent_description": ["a general purpose high-level agent"],
    "model_name": "gpt-4.1",
    "mcp_servers": [
        {
            "addr": "http://localhost:1234/mcp",
            "headers": {
                "Authorization": "Bearer your-token"
            }
        }
    ],
    "sub_agents": [
        <recursive>
    ]
}
```

### models.json

Define available models:

```json
{
    "models": {
        "gpt-4.1": {
            "url": "https://api.example.com/v1/chat/completions",
            "name": "gpt-4.1",
            "key": "your-api-key"
        }
    }
}
```