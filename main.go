package main

import (
	"log/slog"
	"os"

	"github.com/shuv1824/recommender/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}
