package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"jainam-panchal/nextgen-training/text-editor/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := cli.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
