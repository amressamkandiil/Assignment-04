package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type Client struct {
	id   int64
	conn net.Conn
	send chan string
}

var (
	clients     = make(map[int64]*Client)
	clientsMu   sync.Mutex
	broadcastCh = make(chan Broadcast)
	nextID      int64
)

type Broadcast struct {
	fromID int64
	text   string
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server listening on :9000")

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		id := atomic.AddInt64(&nextID, 1)
		client := &Client{
			id:   id,
			conn: conn,
			send: make(chan string, 16),
		}

		clientsMu.Lock()
		clients[id] = client
		clientsMu.Unlock()

		broadcastCh <- Broadcast{fromID: id, text: fmt.Sprintf("User [%d] joined", id)}

		go clientWriter(client)

		go clientReader(client)
	}
}

func broadcaster() {
	for b := range broadcastCh {
		clientsMu.Lock()
		for id, c := range clients {
			if id == b.fromID {
				continue
			}
			select {
			case c.send <- b.text:
			default:
				fmt.Printf("dropping message to user %d (send buffer full)\n", id)
			}
		}
		clientsMu.Unlock()
	}
}

func clientWriter(c *Client) {
	defer c.conn.Close()
	writer := bufio.NewWriter(c.conn)
	for msg := range c.send {
		_, err := writer.WriteString(msg + "\n")
		if err != nil {
			fmt.Println("write error to", c.id, err)
			break
		}
		if err := writer.Flush(); err != nil {
			fmt.Println("flush error to", c.id, err)
			break
		}
	}
}

func clientReader(c *Client) {
	defer func() {
		c.conn.Close()
		clientsMu.Lock()
		delete(clients, c.id)
		clientsMu.Unlock()
		broadcastCh <- Broadcast{fromID: c.id, text: fmt.Sprintf("User [%d] left", c.id)}
		close(c.send)
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Text()
		msg := fmt.Sprintf("User [%d]: %s", c.id, line)
		broadcastCh <- Broadcast{fromID: c.id, text: msg}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("read error from", c.id, err)
	}
}
