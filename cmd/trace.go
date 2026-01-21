package cmd

import (
	"fmt"
	"os"
	"time"
)

func tracef(format string, args ...any) {
	if !debugMode {
		return
	}
	ts := time.Now().Format(time.RFC3339Nano)
	fmt.Fprintf(os.Stderr, "debug %s "+format+"\n", append([]any{ts}, args...)...)
}
