package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var mutex sync.Mutex

func exitSaver() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Kill)
	go func() {
		<-c
		os.Create("EXITED")
		os.Exit(1)
	}()
}

func speedTester(ws, ts int) {
	for {
		var tmp Server
		if err := db.Select().Limit(1).OrderBy("-Timestamp").First(&tmp); err != nil {
			fmt.Println(err)
		}

		tmp.Download = float64(ts*8) / SpeedTest(tmp.IP, "download", ws, ts)
		tmp.Upload = float64(ts*8) / SpeedTest(tmp.IP, "upload", ws, ts)
		tmp.Ping = SpeedTest(tmp.IP, "ping", ws, ts) * 1000 / 4
		tmp.Timestamp = time.Now()

		buff, _ := json.Marshal(tmp)
		m.Broadcast(buff)

		fmt.Println(tmp)

		time.Sleep(time.Second)
	}
}

func SpeedTest(ip, mode string, window_size, test_size int) float64 {
	block_size := window_size * 1024

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:1337", ip), 3*time.Second)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("%s|%d|%d", mode, test_size, window_size)))

	t := time.Now()

	switch mode {
	case "download":
		buf := make([]byte, block_size)
		for {
			_, err := conn.Read(buf)

			if err == io.EOF {
				break
			}
		}
	case "upload":
		buf := make([]byte, block_size)
		rand.Read(buf)
		for i := 0; i < test_size*1024*1024/block_size; i++ {
			conn.Write(buf)
		}
	case "ping":
		buf := make([]byte, 64)
		tmp := make([]byte, 64)
		rand.Read(buf)

		for i := 0; i < 4; i++ {
			conn.Write(buf)
			conn.Read(tmp)
			// if !bytes.Equal(buf, tmp) {
			// 	return 0.0
			// }
		}
	}

	return time.Since(t).Seconds()
}
