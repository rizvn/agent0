package util

import (
	"fmt"
	"log/slog"
	"os"
)

func ChannelToStdOut(out <-chan string) {
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
