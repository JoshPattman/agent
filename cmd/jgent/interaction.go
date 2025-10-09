package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JoshPattman/agent"
)

func interactionLoop(a *agent.Agent) {
	// Visual terminal stuff
	a.SetOnReActCompleteCallback(func(ras agent.ReActStep) {
		longTextLimit := 100
		reasoning := ras.Reasoning
		if len(reasoning) > longTextLimit {
			reasoning = reasoning[:longTextLimit-3] + "..."
		}
		reasoning = strings.ReplaceAll(reasoning, "\n", " ")
		fmt.Printf("	\033[33m%s\033[0m\n", reasoning)

		for _, actionObservation := range ras.ActionObservations {
			argsStr := agent.FormatActionArgsForDisplay(actionObservation.Action.Args)
			fmt.Printf("	$ \033[34mtool://%s?%s\033[0m\n",
				actionObservation.Action.Name,
				argsStr,
			)
			obs := actionObservation.Observation.Observed
			if len(obs) > longTextLimit {
				obs = obs[:longTextLimit-3] + "..."
			}
			obs = strings.ReplaceAll(obs, "\n", " ")
			fmt.Printf("	> \033[35m%s\033[0m\n", obs)
		}
	})

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🤖 Chat started — type 'exit' or 'quit' to end.")
	fmt.Println()

	for {
		fmt.Print("\033[36mYou:\033[0m ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			fmt.Println("👋 Goodbye!")
			break
		}

		answer, err := a.Answer(input)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\033[32mAgent:\033[0m %s\n\n", answer)
	}
}
