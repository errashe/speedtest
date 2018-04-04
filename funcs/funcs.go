package Funcs

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

	. "../structs"

	"github.com/asdine/storm"
	"github.com/labstack/echo"
	"gopkg.in/olahol/melody.v1"
	"gopkg.in/unrolled/render.v1"
)

var mutex sync.Mutex

func ExitSaver() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Kill)
	go func() {
		<-c
		os.Create("EXITED")
		os.Exit(1)
	}()
}

func SpeedTester(ws, ts int, m *melody.Melody, db *storm.DB) {
	t := time.Now()
	for {
		var tmp Server
		if err := db.Select().Limit(1).OrderBy("Timestamp").First(&tmp); err == storm.ErrNotFound {
			fmt.Println("Необходимо инициализировать базу")
			os.Exit(0)
		} else if err != nil {
			fmt.Println(err)
			continue
		}

		if tmp.IP == "" {
			fmt.Println("not found")
			return
		}

		tmp.Download = float64(ts*8) / SpeedTest(tmp.IP, "download", ws, ts)
		tmp.Upload = float64(ts*8) / SpeedTest(tmp.IP, "upload", ws, ts)
		tmp.Ping = SpeedTest(tmp.IP, "ping", ws, ts) * 1000 / 4
		tmp.Timestamp = time.Now()

		var h History
		h.Copy(tmp)

		if err := db.Save(&h); err != nil {
			fmt.Println(err)
		}

		buff, _ := json.Marshal(tmp)
		m.Broadcast(buff)
		fmt.Printf("\r%s\n%.2f", tmp, time.Since(t).Seconds())

		if err := db.Save(&tmp); err != nil {
			fmt.Println(err)
		}

		time.Sleep(time.Second)
	}
}

func SpeedTest(ip, mode string, window_size, test_size int) float64 {
	block_size := window_size * 1024
	t := time.Now()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:1337", ip), 3*time.Second)
	if err != nil {
		fmt.Println(err)
		return 10000.0
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("%s|%d|%d", mode, test_size, window_size)))

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

func NewRender(or render.Options) *RenderWrapper {
	r := &RenderWrapper{}
	r.SetRender(render.New(or))
	return r
}

type RenderWrapper struct { // We need to wrap the renderer because we need a different signature for echo.
	rnd *render.Render
}

func (r *RenderWrapper) SetRender(ir *render.Render) {
	r.rnd = ir
}

func (r *RenderWrapper) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.rnd.HTML(w, 0, name, data) // The zero status code is overwritten by echo.
}
