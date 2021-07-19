package main

import (
	"bufio"
	"fmt"
	"lite_chat/util"
	"net"
	"os"
	"strings"
	"time"
)

const (
	AUTHOR="huangjw_bupt@qq.com"
	LiteChat = "lite-chat"
	GoodByeMsg = "HAVE A NICE DAY! See u next time!\n"
)

func main() {
	//ip, port := "localhost", "8080"
	ip, port := "123.57.29.103", "8080"
	host := ip + ":" + port
	fmt.Println("Welcome to lite-chat!\nYou can input 'bye' to quit anytime and input 'exit' to switch other rooms")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Printf("\nSorry, something went wrong...\n" +
			"lite-chat will be more awesome if u can send the err to %s\n", AUTHOR)
		return
	}
	fmt.Println("connected ")
	defer func() {
		conn.Close()
		sayBye()
		time.Sleep(10 * time.Second)
	}()
	go func(conn net.Conn) {
		_ = util.MustCopy(os.Stdout, conn)
	}(conn)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if text == "bye" {
			return
		}
		if n, err := fmt.Fprintln(conn, text); err != nil || n == 0 {
			fmt.Printf("\nSorry, something went wrong...\n" +
				"lite-chat will be more awesome if u can send the err to %s\n", AUTHOR)
			return
		}
	}
}

func sayBye() {
	var spaces [len(GoodByeMsg)]string
	fmt.Printf("\n%s\n%s——%s\n", GoodByeMsg, strings.Join(spaces[:], " "), LiteChat)
}
