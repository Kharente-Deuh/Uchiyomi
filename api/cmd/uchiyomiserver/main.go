// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	isHealthcheck := flag.Bool("healthcheck", false, "probe /readyz locally and exit 0 when the server is ready")
	flag.Parse()

	if *isHealthcheck {
		os.Exit(runHealthcheck())
	}

	cfg, err := newConfig()
	if err != nil {
		log.Fatalf("invalid configuration: %s", err.Error())
	}

	app, err := setupApp(cfg)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "app.Run: %s\n", err)
		os.Exit(1)
	}
}
