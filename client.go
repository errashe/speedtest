package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	l, err := net.Listen("tcp", ":1337")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	rand.Seed(time.Now().Unix())

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)

	strs := strings.Split(strings.TrimSpace(string(buf[:n])), "|")
	test_size, _ := strconv.Atoi(strs[1])
	WINDOW_SIZE, _ := strconv.Atoi(strs[2])
	BLOCK_SIZE := WINDOW_SIZE * 1024

	buf = make([]byte, BLOCK_SIZE)

	switch string(strs[0]) {
	case "download":
		fmt.Println("DOWNLOADING")
		rand.Read(buf)
		for i := 0; i < test_size*1024*1024/BLOCK_SIZE; i++ {
			conn.Write(buf)
		}
	case "upload":
		fmt.Println("UPLOADING")
		for {
			_, err := conn.Read(buf)

			if err == io.EOF {
				break
			}
		}
	}
}
