package main

import (
	"bufio"
	"fmt"
	"lite_chat/util"
	"net"
	"os"
)

func main() {
	ip, port := "47.101.134.245", "8081"
	host := ip + ":" + port
	fmt.Println("lite-chat client is connecting to server=" + host)
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println("connect err: ", err)
	}
	fmt.Println("connected ")
	defer func() {
		fmt.Println("client sign out, closing connection...")
		conn.Close()
	}()
	go func(conn net.Conn) {
		for {
			if err := util.MustCopy(os.Stdout, conn); err != nil {
				//fmt.Println("copy err: ", err)
				return
			}
		}
	}(conn)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if text == "bye" {
			return
		}
		if _, err := fmt.Fprintln(conn, text); err != nil {
			fmt.Println(err)
			return
		}
	}
}

