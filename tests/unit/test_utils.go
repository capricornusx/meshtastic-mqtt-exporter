package tests

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func findFreePort(t *testing.T) int {
	return findFreePortExcluding(t, 0)
}

func findFreePortExcluding(t *testing.T, excludePort int) int {
	for i := 0; i < 20; i++ {
		// #nosec G102 - Binding to all interfaces is needed for testing
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			continue
		}
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		if port == excludePort {
			continue
		}

		// Проверяем, что порт действительно свободен
		time.Sleep(10 * time.Millisecond)
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			// Порт свободен
			return port
		}
		conn.Close()
	}
	t.Fatal("could not find free port after 20 attempts")
	return 0
}

func waitForHTTPServer(t *testing.T, port int) {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	const httpTimeout = 500 * time.Millisecond
	client := &http.Client{Timeout: httpTimeout}

	for i := 0; i < 50; i++ {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("HTTP server did not start on port %d within 5s", port)
}
