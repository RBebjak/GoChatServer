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
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run client.go <server:port> <username>")
		return
	}

	addr := os.Args[1]
	username := os.Args[2]

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

	console := bufio.NewScanner(os.Stdin)
	for console.Scan() {
		room := strings.TrimSpace(console.Text())

		if room == "exit" {
			fmt.Printf("{Cant use %s as name of room}\n", room)
			continue
		}
		if room == "logout" {
			return
		}

		// Send room
		fmt.Fprintf(conn, "%s\n", room)
		fmt.Println(readLine(reader))

		stopFetch := make(chan struct{})

		go func() {
			for {
				select {
				case <-stopFetch:
					return
				case <-time.After(2 * time.Second):
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
				}
			}
		}()

		// Read user input from console
		for room != "" && console.Scan() {
			text := strings.TrimSpace(console.Text())
			if text == "" {
				continue
			}

			fmt.Fprintf(conn, "%s\n", text)

			if text == "exit" {
				close(stopFetch)
				room = ""
				fmt.Println(readLine(reader))
				break
			}

			if text == "logout" {
				close(stopFetch)
				return
			}

			// Print server response
			fmt.Println(readLine(reader))
		}
	}
}

func readLine(reader *bufio.Reader) string {
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}
