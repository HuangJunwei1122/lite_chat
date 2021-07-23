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
		printWrong()
		return
	}
	fmt.Println("connected ")
	defer func() {
		conn.Close()
		sayBye()
		time.Sleep(10 * time.Second)
	}()
	go func(conn net.Conn) {
		//_ = util.MustCopy(os.Stdout, conn)
		_ = util.PrintStdout(conn)
	}(conn)
	input := bufio.NewReader(os.Stdin)
	for {
		text, err := input.ReadString('\n')
		if err != nil {
			printWrong()
			return
		}
		text = text[:len(text) - 1]
		if text == "bye" {
			return
		}
		if n, err := fmt.Fprintln(conn, text); err != nil || n == 0 {
			printWrong()
			return
		}
	}
}

func sayBye() {
	var spaces [len(GoodByeMsg)]string
	fmt.Printf("\n%s\n%s——%s\n", GoodByeMsg, strings.Join(spaces[:], " "), LiteChat)
}

func printWrong() {
	fmt.Printf("\nSorry, something went wrong...\n" +
		"lite-chat will be more awesome if u can send the err to %s\n", AUTHOR)
}

