package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	RoomNumLimit    = 100
	RoomPlayerLimit = 100

	MsgEnterOk  = 0
	MsgRoomFull = 1
	MsgNoRoom   = 2
)

type Room struct {
	ID       int
	members  map[int]*Client
	entering chan *Client
	leaving  chan *Client
	messages chan string
	maxMID   int
}

type Client struct {
	Conn    *net.Conn
	MsgChan chan string
	name    string
	ID      int
}

type EnterMsg struct {
	RID      int
	Member   *Client
	Resp     chan int
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
	}
)

func tryEnterRoom(client *Client, rid int) int {
	room, ok := rooms[rid]
	if !ok {
		if len(rooms) < RoomNumLimit {
			newRoom := Room{
				ID: rid,
				members: make(map[int]*Client),
				entering: make(chan *Client, RoomPlayerLimit),
				leaving: make(chan *Client, RoomPlayerLimit),
				messages: make(chan string, RoomPlayerLimit),
			}
			go runRoom(&newRoom)
			newRoom.entering <- client
			rooms[rid] = &newRoom
			return MsgEnterOk
		} else {
			return MsgNoRoom
		}
	} else {
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
			msg.Resp <- tryEnterRoom(msg.Member, msg.RID)
			fmt.Printf("create room=%d, available rooms=%v\n", msg.RID, rooms)
		case rid := <- leaving:
			delete(rooms, rid)
			fmt.Printf("close room=%d, available rooms=%v\n", rid, rooms)
		case <-closing:
			break
		default:
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
			room.messages <- fmt.Sprintf("%d-%s entered.", client.ID, client.name)
		case client := <-room.leaving:
			delete(room.members, client.ID)
			clientID := client.ID
			client.ID = 0
			if len(room.members) == 0 {
				leaving <- room.ID
				return
			}
			room.messages <- fmt.Sprintf("%d-%s left.", clientID, client.name)
		}
	}
}

func handleClient(client *Client) {
	conn := *client.Conn
	defer func() {
		close(client.MsgChan)
		conn.Close()
		fmt.Printf("connection lost: %s\n", conn.RemoteAddr().String())
	}()
	go writeConn(client)

	// try login
	err := login(client)
	if err != nil {
		fmt.Printf("handleClient err, login fail, client=%s-%s, err=%s\n",
			client.name, conn.RemoteAddr().String(), err.Error())
	}

	// already login
	for {
		room := enterRoom(client)
		if room == nil {
			break
		}
		who := fmt.Sprintf("%d-%s", client.ID, client.name)
		input := bufio.NewScanner(conn)
		for input.Scan() {
			msg := input.Text()
			if msg == "exit"{
				break
			}
			room.messages <- who + ": " + msg
		}
		room.leaving <- client
	}
}

func login(client *Client) error {
	conn := *client.Conn
	for {
		if _, err := io.WriteString(conn, "Tell me your awesome name >>> "); err != nil {
			return err
		}
		n, err := fmt.Fscanln(conn, &client.name)
		if n > 0 && err == nil {
			return nil
		}
		if _, err := io.WriteString(conn, "\nAn awesome name must not be blank.\n"); err != nil {
			return err
		}
	}
}

func enterRoom(client *Client) *Room {
	conn := *client.Conn
	var rid string
	for {
		if _, err := io.WriteString(conn, "Tell me the room(id) you want to go >>> "); err != nil {
			return nil
		}
		_, err := fmt.Fscanln(conn, &rid)
		if err != nil {
			return nil
		}
		if roomID, err2 := strconv.Atoi(rid); err2 != nil {
			_, _ = io.WriteString(conn, "Invalid room ID")
			return nil
		} else {
			resp := make(chan int)
			entering <- &EnterMsg{roomID, client, resp}
			code := <- resp
			if msg, ok := response[code]; ok {
				if _, err := io.WriteString(conn, msg); err != nil {
					return nil
				}
			}
			if room, ok := rooms[roomID]; code == MsgEnterOk && ok {
				return room
			}
		}
	}
}

func writeConn(client *Client) {
	for msg := range client.MsgChan {
		n, err := io.WriteString(*client.Conn, msg)
		if err != nil {
			fmt.Printf("writeConn error, n=%d, who=%s-%s, err=%s",
				n, client.name, (*client.Conn).RemoteAddr().String(), err.Error())
		}
	}
}

func main() {
	port := "8081"
	go handleRoom()
	defer func() {
		close(closing)
	}()
	fmt.Println("tcp server is listening " + port)
	listener, err := net.Listen("tcp", "0.0.0.0:" + port)
	fmt.Println("lite-chat server started")
	if err != nil {
		fmt.Println("fatal err: ", err)
		return
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("accepted client", conn.RemoteAddr().String())
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