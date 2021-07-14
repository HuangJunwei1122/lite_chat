package main

import (
	"bufio"
	"fmt"
	"lite_chat/util"
	"net"
	"os"
)

const (
	AUTHOR="huangjw_bupt@qq.com"
)

func main() {
	ip, port := "localhost", "8080"
	//ip, port := "123.57.29.103", "8080"
	host := ip + ":" + port
	fmt.Println("lite-chat start...")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Printf("Sorry, something went wrong...\n" +
			"lite-chat will be more awesome if u can send the err to %s! See u", AUTHOR)
		return
	}
	fmt.Println("connected ")
	defer func() {
		conn.Close()
	}()
	go func(conn net.Conn) {
		for {
			if err := util.MustCopy(os.Stdout, conn); err != nil {
				return
			}
		}
	}(conn)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		if text == "bye" {
			fmt.Println("HAVE A NICE DAY! See u~\n                      ——lite-chat")
			return
		}
		if n, err := fmt.Fprintln(conn, text); err != nil || n == 0 {
			fmt.Println(err)
			return
		}
	}
}
