package dataplane

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestSendUDS(t *testing.T) {
	socketPath := "/tmp/test_uds_sender.sock"

	// Cleanup if exists
	os.Remove(socketPath)

	// Create a listener
	conn, err := net.ListenPacket("unixgram", socketPath)
	if err != nil {
		t.Fatalf("Failed to create UDS listener: %v", err)
	}
	defer func() {
		conn.Close()
		os.Remove(socketPath)
	}()

	// Channel to signal reception
	received := make(chan []byte, 1)

	go func() {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			t.Logf("ReadFrom error: %v", err)
			return
		}
		received <- buf[:n]
	}()

	// Wait for listener to be ready (approx)
	time.Sleep(100 * time.Millisecond)

	testData := []byte("hello uds")
	err = SendUDS(socketPath, testData)
	if err != nil {
		t.Fatalf("SendUDS failed: %v", err)
	}

	select {
	case data := <-received:
		if string(data) != string(testData) {
			t.Errorf("Expected %s, got %s", testData, data)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for packet")
	}
}

func TestSendUDS_NoSocket(t *testing.T) {
	socketPath := "/tmp/non_existent_socket.sock"
	err := SendUDS(socketPath, []byte("data"))
	if err == nil {
		t.Error("Expected error when sending to non-existent socket, got nil")
	}
}

func TestSendUDS_InvalidPath(t *testing.T) {
	// A directory is an invalid socket path
	err := SendUDS("/tmp", []byte("data"))
	if err == nil {
		t.Error("Expected error when sending to invalid path, got nil")
	}
}
