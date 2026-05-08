package main

import (
	"agent0/agent"
	"agent0/util"
	"flag"
	"log/slog"
)

func main() {
	//read config from env
	c, err := agent.NewConfigFromEnv(".env")
	if err != nil {
		panic(err)
	}

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
