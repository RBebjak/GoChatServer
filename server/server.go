package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type ChatLog struct {
	Messages []Message
	writeCh  chan Message
}

func newChatLog() *ChatLog {
	cl := &ChatLog{
		writeCh: make(chan Message, 100),
	}
	go cl.run()
	return cl
}

func (cl *ChatLog) run() {
	for msg := range cl.writeCh {
		cl.Messages = append(cl.Messages, msg)
	}
}

func (cl *ChatLog) addMessage(user, text string) {
	cl.writeCh <- Message{User: user, Message: text}
}

func (cl *ChatLog) getMessagesSince(lastSeen int, excludeUser string) ([]Message, int) {
	if lastSeen < len(cl.Messages) {
		var result []Message
		for _, msg := range cl.Messages[lastSeen:] {
			if msg.User != excludeUser {
				result = append(result, msg)
			}
		}
		return result, len(cl.Messages)
	}
	return []Message{}, lastSeen
}

var logs = make(map[string]*ChatLog)

func getLog(name string) *ChatLog {
	if logs[name] == nil {
		logs[name] = newChatLog()
	}
	return logs[name]
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewScanner(conn)

	// First line: username
	if !reader.Scan() {
		return
	}
	user := strings.TrimSpace(reader.Text())
	if user == "" {
		fmt.Fprintf(conn, "{\"error\":\"username required\"}\n")
		return
	}

	var chatLog *ChatLog
	lastSeen := 0
	currentRoom := ""

	fmt.Fprintf(conn, "{\"status\":\"hello %s\"}\n", user)

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		switch line {
		case "/fetch":
			if chatLog == nil {
				fmt.Fprintf(conn, "{\"error\":\"not in a room\"}\n")
				continue
			}
			newMsgs, newIndex := chatLog.getMessagesSince(lastSeen, user)
			resp := struct {
				Messages []Message `json:"messages"`
			}{newMsgs}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(conn, "%s\n", data)
			lastSeen = newIndex

		case "exit":
			if chatLog != nil {
				fmt.Fprintf(conn, "{\"status\":\"left room %s\"}\n", currentRoom)
				chatLog = nil
				currentRoom = ""
				lastSeen = 0
			} else {
				fmt.Fprintf(conn, "{\"error\":\"not in a room\"}\n")
			}

		case "logout":
			fmt.Fprintf(conn, "{\"status\":\"bye %s\"}\n", user)
			return

		default:
			if chatLog == nil {
				// Treat as room name to join
				room := line
				chatLog = getLog(room)
				currentRoom = room
				lastSeen = 0
				fmt.Fprintf(conn, "{\"status\":\"joined room %s\"}\n", room)
			} else {
				// Normal message
				chatLog.addMessage(user, line)
				fmt.Fprintf(conn, "{\"status\":\"sent\"}\n")
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("TCP chat server running on :8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}
