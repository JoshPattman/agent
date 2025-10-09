package main

import "time"

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
