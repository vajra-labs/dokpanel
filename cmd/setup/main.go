package main

import (
	"context"
	"fmt"
	"os"

	"dokpanel/src/docker"
	_ "dokpanel/src/logger"
	"dokpanel/src/setup"
)

func main() {
	docker.Init()
	ctx := context.Background()

	if len(os.Args) > 1 && os.Args[1] == "--teardown" {
		fmt.Println("Starting dokpanel teardown...")
		if err := setup.RunTeardown(ctx); err != nil {
			fmt.Printf("Teardown failed: %v\n", err)
			return
		}
		fmt.Println("Teardown completed successfully!")
		return
	}

	fmt.Println("Starting dokpanel setup...")
	if err := setup.RunSetup(ctx); err != nil {
		fmt.Printf("Setup failed: %v\n", err)
		return
	}
	fmt.Println("Setup completed successfully!")
}
