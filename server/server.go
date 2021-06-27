package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

type Client struct {
	Conn    *net.Conn
	MsgChan chan string
}

var (
	clients = make(map[*Client]bool)
	entering = make(chan *Client)
	leaving = make(chan *Client)
	messages = make(chan string)
)

func main() {
	go broadcast()
	fmt.Println("tcp server is listening 8080")
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	fmt.Println("lite-chat server started")
	if err != nil {
		fmt.Println("fatal err: ", err)
		return
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("accepted client", conn)
		if err != nil {
			fmt.Println("accept err: ", err)
			continue
		}
		go handleClient(&Client{
			Conn:    &conn,
			MsgChan: make(chan string),
		})
	}
}

func broadcast() {
	for {
		select {
		case msg := <-messages:
			for client := range clients {
				client.MsgChan <- msg + "\n"
			}
		case client := <-entering:
			clients[client] = true
		case client := <-leaving:
			delete(clients, client)
			close(client.MsgChan)
		}
	}
}

func handleClient(client *Client) {
	conn := *client.Conn
	who := conn.RemoteAddr().String()
	entering <- client
	messages <- who + " entered."
	go writeConn(client)
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ": " + input.Text()
	}
	leaving <- client
	conn.Close()
	messages <- who + " left."
	fmt.Println("connection lost: ", who)
}

func writeConn(client *Client) {
	for msg := range client.MsgChan {
		io.WriteString(*client.Conn, msg)
	}
}
