package main

import (
	"io"
	. "log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var buf []byte

func main() {
	l, err := net.Listen("tcp", ":1337")
	if err != nil {
		Println(err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			Println(err)
			continue
		}

		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	rand.Seed(time.Now().Unix())

	buf = make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		Printf("%x\n", err)
		return
	}

	strs := strings.Split(strings.TrimSpace(string(buf[:n])), "|")
	test_size, err := strconv.Atoi(strs[1])
	if err != nil {
		Println(err)
		return
	}
	WINDOW_SIZE, err := strconv.Atoi(strs[2])
	if err != nil {
		Println(err)
		return
	}
	BLOCK_SIZE := WINDOW_SIZE * 1024

	switch string(strs[0]) {
	case "download":
		tmp := make([]byte, BLOCK_SIZE)
		rand.Read(tmp)
		for i := 0; i < test_size*1024*1024/BLOCK_SIZE; i++ {
			_, err := conn.Write(tmp)
			if err != nil {
				Println(err)
				return
			}
		}
	case "upload":
		for {
			tmp := make([]byte, BLOCK_SIZE)
			_, err := conn.Read(tmp)
			if err == io.EOF {
				break
			} else if err != nil {
				Println(err)
				return
			}
		}
	case "ping":
		for {
			tmp := make([]byte, 64)
			_, err := conn.Read(tmp)
			if err == io.EOF {
				break
			} else if err != nil {
				Println(err)
				return
			}

			conn.Write(tmp)
		}
	}
}
