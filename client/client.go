package main

import (
	"bufio"
	"fmt"
	"lite_chat/util"
	"net"
	"os"
)

func main() {
	//ip, port := "localhost", "8080"
	ip, port := "123.57.29.103", "8080"
	host := ip + ":" + port
	fmt.Println("lite-chat client is connecting to server")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println("connect err: ", err)
		return
	}
	fmt.Println("connected ")
	defer func() {
		fmt.Println("client sign out, closing connection...")
		conn.Close()
	}()
	go func(conn net.Conn) {
		for {
			if err := util.MustCopy(os.Stdout, conn); err != nil {
				fmt.Println("MustCopy err", err)
			}
		}
	}(conn)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if text == "bye" {
			fmt.Println("test")
			return
		}
		if n, err := fmt.Fprintln(conn, text); err != nil || n == 0 {
			fmt.Println(err)
			return
		}
	}
}
