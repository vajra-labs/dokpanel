package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dokpanel/src"
	"dokpanel/src/conf"
	"dokpanel/src/db"
	"dokpanel/src/docker"
	_ "dokpanel/src/logger"

	"github.com/gofiber/fiber/v3"
)

func main() {
	docker.Init()

	app := src.App()
	uri := fmt.Sprintf("%s:%d", conf.Env.HOST, conf.Env.PORT)

	// Listen from a different goroutine
	if err := app.Listen(uri, fiber.ListenConfig{
		EnablePrefork: false,
	}); err != nil && errors.Is(err, net.ErrClosed) {
		log.Panic(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(done)

	<-done // This blocks the main thread until an interrupt is received
	fmt.Println("\nGracefully shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		fmt.Printf("Server shutdown failed: %v\n", err)
	}

	// Your cleanup tasks go here
	fmt.Println("Running cleanup tasks...")
	db.Close()
	docker.Close()
	fmt.Println("Cleanup completed. Bye 👋")
}
