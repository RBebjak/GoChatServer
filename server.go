package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Message struct {
	User string
	Text string
}

type MessageNode struct {
	Msg  Message
	Next *MessageNode
}

var (
	chats   = make(map[string][]Message)
	chatsMu sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	fmt.Println("Client connected:", conn.RemoteAddr())

	if scanner.Scan() {
		text := scanner.Text()
		if text != "Connect" {
			conn.Write([]byte("Invalid start\n"))
			return
		}
		conn.Write([]byte("UserName\n"))
	} else {
		return
	}

	if scanner.Scan() {
		scanner.Text()
		conn.Write([]byte("UserPassword\n"))
	} else {
		return
	}

	if scanner.Scan() {
		scanner.Text()
		conn.Write([]byte("Welcome!\n"))
	} else {
		return
	}

	fmt.Printf("User logged in\n")

	chat := ""
	if scanner.Scan() {
		chat = scanner.Text()
	}

	chatMsgs := make(chan string)
	msg := <-chatMsgs
	msg += chat
	fmt.Println(msg)
}

func main() {
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error Accept client:", err)
			continue
		}

		go handleConnection(conn)
	}

}
