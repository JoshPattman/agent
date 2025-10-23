package ai

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type timeTool struct {
}

func (t *timeTool) Call(map[string]any) (string, error) {
	return time.Now().Format(time.ANSIC), nil
}

func (t *timeTool) Name() string {
	return "get_time"
}

func (t *timeTool) Description() []string {
	return []string{
		"Gets the current time",
		"Takes no arguments",
	}
}

type listDirectoryTool struct {
}

func (t *listDirectoryTool) Call(args map[string]any) (string, error) {
	pathRaw, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("missing required argument: path")
	}
	path, ok := pathRaw.(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}

	// If path is empty, use current directory
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %v", path, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of directory: %s\n", path))
	result.WriteString("Type\tName\n")
	result.WriteString("----\t----\n")

	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir() {
			entryType = "dir"
		}
		result.WriteString(fmt.Sprintf("%s\t%s\n", entryType, entry.Name()))
	}

	return result.String(), nil
}

func (t *listDirectoryTool) Name() string {
	return "list_directory"
}

func (t *listDirectoryTool) Description() []string {
	return []string{
		"Lists the contents of a directory (like ls command)",
		"Takes one argument: 'path' (string) - the directory path to list",
		"If path is empty or not provided, lists current directory",
	}
}

type readFileTool struct {
}

func (t *readFileTool) Call(args map[string]any) (string, error) {
	pathRaw, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("missing required argument: path")
	}
	path, ok := pathRaw.(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return string(content), nil
}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Description() []string {
	return []string{
		"Reads and returns the entire content of a file",
		"Takes one argument: 'path' (string) - the file path to read",
	}
}
