package main

import (
	"context"
	"fmt"
	"os"

	"dokpanel/src/conf"
	"dokpanel/src/lib/docker"
	"dokpanel/src/logger"
	"dokpanel/src/setup"

	"go.uber.org/fx"
)

func run(runner *setup.Runner) {
	ctx := context.Background()
	teardown := len(os.Args) > 1 && os.Args[1] == "--teardown"

	if teardown {
		fmt.Println("Starting dokpanel teardown...")
		if err := runner.Teardown(ctx); err != nil {
			fmt.Printf("Teardown failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Teardown completed successfully!")
		return
	}

	fmt.Println("Starting dokpanel setup...")
	if err := runner.Setup(ctx); err != nil {
		fmt.Printf("Setup failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Setup completed successfully!")
}

func main() {
	app := fx.New(
		fx.NopLogger,
		conf.Module,
		logger.Module,
		docker.Module,
		setup.Module,
		fx.Invoke(run),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Printf("Startup failed: %v\n", err)
		os.Exit(1)
	}
	defer app.Stop(ctx) //nolint:errcheck
}
