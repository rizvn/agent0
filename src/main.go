package main

import (
	"agent0/agent"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	//read config from env
	c, err := agent.NewConfigFromEnv(".env")
	if err != nil {
		panic(err)
	}

	configureLogging(c.LogLevel)

	// create agent
	a := agent.NewAgent(c)

	// read prompt from if provided on cli
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	// if prompt provided on cli
	if prompt != "" {
		// request resonse mode
		response, err := a.DirectResponse(context.TODO(), prompt)
		if err != nil {
			slog.Error(err.Error(), "err", err)
		}
		fmt.Println(response)
	} else {
		// interactive mode
		err = a.Loop()
		if err != nil {
			panic(err)
		}
	}
}

// configureLogging sets the global logger based on the provided log level
func configureLogging(logLevel string) {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{
		Level: level,
	}

	logger := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(logger))
}
