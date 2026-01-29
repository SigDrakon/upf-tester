package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"upftester/internal/config"
	"upftester/internal/handler"
	"upftester/internal/network"
)

const TestCasePath = "../testcases/complete_test_case/complete_test_case.yaml"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var config config.Config
	err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		log.Fatal(err)
		return
	}

	udpTransport, err := network.NewUDPTransport(config.Basic.LocalN4Ip, "8805", config.Resource.QueueSize)
	if err != nil {
		log.Fatal(err)
		return
	}
	udpTransport.Start()
	defer udpTransport.Stop()

	handler.NewPFCPDispatcher(udpTransport).Start()

	err = handler.SendPFCPAssociationRequest(config.Basic.LocalN4Ip, config.Basic.UpfN4Ip, udpTransport)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Load test cases
	if len(config.TestCases) > 0 {
		for _, path := range config.TestCases {
			log.Printf("Loading test case from: %s", path)
			handler.LoadTestCases(path, &handler.GlobalTestCases)
		}
	} else {
		log.Println("No test cases in config, loading default test case")
		handler.LoadTestCases(TestCasePath, &handler.GlobalTestCases)
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", config.Basic.UpfN4Ip+":8805")
	if err != nil {
		log.Fatal(err)
		return
	}
	handler.RunTestCases(remoteAddr, udpTransport)

	log.Println("Test cases completed")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	log.Println("Received shutdown signal, exiting...")
}
