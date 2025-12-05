package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <name> OR just connect without name, e.g. go run client.go")
	}
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("server read error:", err)
		}
		os.Exit(0)
	}()

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		line := stdin.Text()
		_, err := fmt.Fprintln(conn, line)
		if err != nil {
			fmt.Println("send error:", err)
			return
		}
	}
	if err := stdin.Err(); err != nil {
		fmt.Println("stdin error:", err)
	}
}
