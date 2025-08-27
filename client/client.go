package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run client.go <server:port> <username> <room>")
		return
	}

	addr := os.Args[1]
	username := os.Args[2]
	room := os.Args[3]

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Send username
	fmt.Fprintf(conn, "%s\n", username)
	fmt.Println(readLine(reader))

	// Send room
	fmt.Fprintf(conn, "%s\n", room)
	fmt.Println(readLine(reader))

	// Channel to stop fetch goroutine
	stopFetch := make(chan struct{})

	go func() {
		lastSeen := 0
		for {
			select {
			case <-stopFetch:
				return
			case <-time.After(2 * time.Second):
				// send /fetch
				fmt.Fprintf(conn, "/fetch\n")
				line := readLine(reader)
				var resp struct {
					Messages []Message `json:"messages"`
				}
				if err := json.Unmarshal([]byte(line), &resp); err == nil {
					for _, msg := range resp.Messages {
						fmt.Printf("[%s] %s\n", msg.User, msg.Message)
					}
				}
				lastSeen++ // we are only fetching; server handles lastSeen per connection
			}
		}
	}()

	// Read user input from console
	console := bufio.NewScanner(os.Stdin)
	for console.Scan() {
		text := strings.TrimSpace(console.Text())
		if text == "" {
			continue
		}

		fmt.Fprintf(conn, "%s\n", text)

		// Stop fetch goroutine if client exits room or logs out
		if text == "exit" || text == "logout" {
			close(stopFetch)
			break
		}

		// Print server response
		fmt.Println(readLine(reader))
	}
}

func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}
