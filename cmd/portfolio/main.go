package main

import (
	"log"
	"os"
	"sync"

	"portfolio-server/internal/data"
	"portfolio-server/internal/discord"
	"portfolio-server/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	_ = godotenv.Load()

	// 2. Initialize Discord Client
	dc := discord.New()

	// 3. Load Data Store
	store, err := data.LoadAll(".")
	if err != nil {
		log.Fatalf("[main] failed to load data: %v", err)
	}
	log.Printf("[main] loaded profile: %s", store.Profile.Name)
	log.Printf("[main] loaded %d experiences", len(store.Profile.Experience))
	log.Printf("[main] loaded %d projects", len(store.Profile.Projects))

	// 4. Start Servers
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222"
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		server.StartHTTP(":"+httpPort, store, dc)
	}()

	go func() {
		defer wg.Done()
		server.StartSSH(":"+sshPort, store, dc)
	}()

	wg.Wait()
}
