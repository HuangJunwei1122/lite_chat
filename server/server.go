package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	RoomNumLimit    = 100
	RoomPlayerLimit = 100

	MsgEnterOk  = 0
	MsgRoomFull = 1
	MsgNoRoom   = 2
	MsgPWFail = 3

	EndDelim = '\t'
)

type Room struct {
	ID        int
	members   map[int]*Client
	entering  chan *Client
	leaving   chan *Client
	messages  chan string
	maxMID    int
	PassWord  string
}

type Client struct {
	Conn    *net.Conn
	MsgChan chan string
	name    string
	ID      int
}

type EnterMsg struct {
	RID       int
	Member    *Client
	Resp      chan int
	PassWord  string
}

var (
	rooms    = make(map[int]*Room)
	entering = make(chan *EnterMsg, RoomNumLimit*RoomPlayerLimit)
	leaving  = make(chan int, RoomNumLimit*RoomPlayerLimit)
	closing  = make(chan struct{})
	response = map[int]string{
		MsgEnterOk:  "Enter room success.\n",
		MsgNoRoom:   "There is no room you are looking for, try other rooms.\n",
		MsgRoomFull: "There is no vacancy in this room.\n",
		MsgPWFail: "Incorrect password.\n",
	}
)

func tryEnterRoom(client *Client, rid int, passwd string) int {
	room, ok := rooms[rid]
	if !ok {
		if len(rooms) < RoomNumLimit {
			newRoom := Room{
				ID: rid,
				members: make(map[int]*Client),
				entering: make(chan *Client, RoomPlayerLimit),
				leaving: make(chan *Client, RoomPlayerLimit),
				messages: make(chan string, RoomPlayerLimit),
				PassWord: passwd,
			}
			go runRoom(&newRoom)
			newRoom.entering <- client
			rooms[rid] = &newRoom
			log.Printf("create room=%d, available rooms=%v\n", rid, rooms)
			return MsgEnterOk
		} else {
			return MsgNoRoom
		}
	} else {
		if room.PassWord != passwd {
			return MsgPWFail
		}
		if len(room.members) < RoomPlayerLimit {
			room.entering <- client
			return MsgEnterOk
		} else {
			return MsgRoomFull
		}
	}
}

func handleRoom() {
	for {
		select {
		case msg := <- entering:
			msg.Resp <- tryEnterRoom(msg.Member, msg.RID, msg.PassWord)
		case rid := <- leaving:
			delete(rooms, rid)
			log.Printf("close room=%d, available rooms=%v\n", rid, rooms)
		case <-closing:
			break
		}
	}
}

func runRoom(room *Room) {
	for {
		select {
		case msg := <-room.messages:
			for _, member := range room.members {
				member.MsgChan <- msg + "\n"
			}
		case client := <- room.entering:
			room.maxMID += 1
			room.members[room.maxMID] = client
			client.ID = room.maxMID
			room.messages <- fmt.Sprintf("%s%d entered.", client.name, client.ID)
		case client := <-room.leaving:
			delete(room.members, client.ID)
			clientID := client.ID
			client.ID = 0
			if len(room.members) == 0 {
				leaving <- room.ID
				return
			}
			room.messages <- fmt.Sprintf("%s%d left.", client.name, clientID)
		}
	}
}

func handleClient(client *Client) {
	conn := *client.Conn
	defer func() {
		close(client.MsgChan)
		conn.Close()
		log.Printf("connection lost: %s\n", conn.RemoteAddr().String())
	}()
	go writeConn(client)

	// try login
	err := login(client)
	if err != nil {
		log.Printf("handleClient err, login fail, client=%s-%s, err=%s\n",
			client.name, conn.RemoteAddr().String(), err.Error())
	}

	// already login
	for {
		room, err := enterRoom(client)
		if err != nil {
			return
		}
		who := fmt.Sprintf("%s%d", client.name, client.ID)
		input := bufio.NewScanner(conn)
		for input.Scan() {
			msg := input.Text()
			if msg == "exit"{
				break
			}
			room.messages <- who + ": " + msg
		}
		room.leaving <- client
		if _, err := writeStringWithEnd(conn, "\nyou are leaving room.\n"); err != nil {
			return
		}
	}
}

func login(client *Client) error {
	conn := *client.Conn
	for {
		if _, err := writeStringWithEnd(conn, "Tell me your awesome name >>> "); err != nil {
			return err
		}
		n, err := fmt.Fscanln(conn, &client.name)
		if n > 0 && err == nil {
			return nil
		}
		if _, err := writeStringWithEnd(conn, "\nAn awesome name must not be blank.\n"); err != nil {
			return err
		}
	}
}

func enterRoom(client *Client) (*Room, error) {
	conn := *client.Conn
	var rid, passwd string
	for {
		if _, err := writeStringWithEnd(conn, "room id >>> "); err != nil {
			//return nil, err
			continue
		}
		_, err := fmt.Fscanln(conn, &rid)
		if err != nil {
			if _, err = writeStringWithEnd(conn, "\nRoom ID can't be empty.\n"); err != nil {
				return nil, err
			}
			continue
		}
		roomID, err := strconv.Atoi(rid)
		if err != nil {
			if _, err = writeStringWithEnd(conn, "\nInvalid room ID.\n"); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := writeStringWithEnd(conn, "room password >>> "); err != nil {
			return nil, err
		}
		if _, err = fmt.Fscanln(conn, &passwd); err != nil {
			if _, err1 := writeStringWithEnd(conn, "\nPassword can't be empty.\n"); err1 != nil {
				return nil, err1
			}
			continue
		}
		resp := make(chan int)
		entering <- &EnterMsg{roomID, client, resp, passwd}
		code := <- resp
		if msg, ok := response[code]; ok {
			if _, err := writeStringWithEnd(conn, msg); err != nil {
				return nil, err
			}
		}
		if room, ok := rooms[roomID]; code == MsgEnterOk && ok {
			return room, nil
		}

	}
}

func writeConn(client *Client) {
	for msg := range client.MsgChan {
		n, err := writeStringWithEnd(*client.Conn, msg)
		if err != nil {
			log.Printf("writeConn error, n=%d, who=%s-%s, err=%s",
				n, client.name, (*client.Conn).RemoteAddr().String(), err.Error())
		}
	}
}

func writeStringWithEnd(conn net.Conn, msg string) (int, error) {
	return io.WriteString(conn, fmt.Sprintf("%s\t", msg))
}

func main() {
	port := "8080"
	go handleRoom()
	defer func() {
		close(closing)
	}()
	log.Println("tcp server is listening " + port)
	listener, err := net.Listen("tcp", "0.0.0.0:" + port)
	log.Println("lite-chat server started")
	if err != nil {
		log.Println("fatal err: ", err)
		return
	}
	for {
		conn, err := listener.Accept()
		log.Println("accepted client", conn.RemoteAddr().String())
		if err != nil {
			log.Println("accept err: ", err)
			continue
		}
		go handleClient(&Client{
			Conn:    &conn,
			MsgChan: make(chan string),
		})
	}
}