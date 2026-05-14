package main

import (
	"agent0/agent"
	"agent0/util"
	"flag"
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
		if err := requestResponse(a, prompt); err != nil {
			slog.Error(err.Error(), "err", err)
		}
	} else {
		// interactive mode
		err = a.Loop()
		if err != nil {
			panic(err)
		}
	}
}

func requestResponse(a *agent.Agent, prompt string) error {
	out := make(chan string)
	util.ChannelToStdOut(out)

	if err := a.GenerateResponse(prompt, out, true); err != nil {
		return util.NewErr("Unable to generate response", err)
	}
	return nil
}

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
