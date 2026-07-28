package main

import (
	"context"
	"log/slog"
	"os"

	"tinyschool-api/internal/command"
)

func main() {
	if err := command.Execute(context.Background()); err != nil {
		slog.Error("Tiny School API stopped", "error", err)
		os.Exit(1)
	}
}
