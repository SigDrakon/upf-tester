package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	socketPath := flag.String("socket", "/tmp/upf_tester_uds.sock", "Path to the Unix Domain Socket")
	flag.Parse()

	// Clean up existing socket
	if _, err := os.Stat(*socketPath); err == nil {
		if err := os.Remove(*socketPath); err != nil {
			log.Fatalf("Failed to remove existing socket: %v", err)
		}
	}

	// Create Unix Domain Socket listener (packet-oriented for unixgram)
	conn, err := net.ListenPacket("unixgram", *socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on socket %s: %v", *socketPath, err)
	}
	defer conn.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		conn.Close()
		os.Remove(*socketPath)
		os.Exit(0)
	}()

	fmt.Printf("Listening on %s (unixgram)...\n", *socketPath)

	buffer := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			// Check if error is due to closed connection (shutdown)
			select {
			case <-sigChan:
				return
			default:
				log.Printf("Read error: %v", err)
				continue
			}
		}

		fmt.Printf("Received %d bytes from %v: %s\n", n, addr, string(buffer[:n]))
	}
}
