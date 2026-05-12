package util

import (
	"fmt"
	"log/slog"
	"os"
)

// ChannelToStdOut writes the contents of a channel to stdout.
// - out is the channel to read from. It is expected that the channel will be closed when done,
// and this function will return at that point.
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
