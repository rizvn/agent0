package agent

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
)

func TestAgent(t *testing.T) {
	c, err := NewConfigFromEnv("Agent_test.env")
	if err != nil {
		t.Log(err)
		log.Default().Print("Skipping test")
		return
	}

	a := NewAgent(c)

	t.Run("Test prompt", func(t *testing.T) {
		prompt := "how many files are there in current directory"
		out := make(chan string)
		printOutChannel(out)

		err = a.GenerateResponse(prompt, out, false)
		if err != nil {
			panic(err)
		}
	})

	t.Run("Test prompt2", func(t *testing.T) {
		prompt := "show contents of file which contains the word agent in current directory"
		out := make(chan string)
		printOutChannel(out)

		err = a.GenerateResponse(prompt, out, false)
		if err != nil {
			panic(err)
		}
	})

	t.Run("Test streaming intermediate", func(t *testing.T) {
		prompt := "show contents of file which contains the word agent in current directory"
		out := make(chan string)
		printOutChannel(out)

		err = a.GenerateResponse(prompt, out, true)
		if err != nil {
			panic(err)
		}
	})

	t.Run("Upsert datasource test", func(t *testing.T) {
		prompt := "Create a new datasource named test with url postgress://test:testpw@localhost:5431/testdb "
		out := make(chan string)
		printOutChannel(out)
		err = a.GenerateResponse(prompt, out, true)
		if err != nil {
			t.Fatal("Expected no errors but got", err)
		}
	})

}

func printOutChannel(out <-chan string) {
	// write output streamNextMessage
	go func() {
		for s := range out {
			_, err := fmt.Fprint(os.Stdout, s)
			if err != nil {
				slog.Error("Unable to write to stdout", "err", err)
			}
		}
	}()
}
