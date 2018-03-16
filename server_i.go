package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/labstack/echo"
	"gopkg.in/olahol/melody.v1"
	"gopkg.in/unrolled/render.v1"
)

type Server struct {
	ID        int       `db:"id"`
	IP        string    `db:"ip"`
	Download  float64   `db:"download"`
	Upload    float64   `db:"upload"`
	Timestamp time.Time `db:"timestamp"`
}

var servers []Server = []Server{
	// Server{ID: 1, IP: "10.14.14.46"},
	// Server{ID: 2, IP: "10.14.15.1"},
	// Server{ID: 3, IP: "10.14.78.30"},
	// Server{ID: 4, IP: "10.14.89.10"},
	// Server{ID: 5, IP: "10.14.51.26"},
	// Server{ID: 123, IP: "10.14.15.1"},
	// Server{ID: 124, IP: "10.14.13.100"},
	Server{ID: 1, IP: "localhost"},
}

func main() {
	window_size := flag.Int("w", 64, "size of window, which one gonna be send")
	test_size := flag.Int("t", 40, "megabytes of data, which one gonna be send")

	flag.Parse()

	m := melody.New()

	go func() {
		for {
			for _, server := range servers {
				server.Download = SpeedTest(server.IP, "download", *window_size, *test_size)
				server.Upload = SpeedTest(server.IP, "upload", *window_size, *test_size)
				server.Timestamp = time.Now()

				buff, _ := json.Marshal(server)
				m.Broadcast(buff)

				fmt.Println(server)

				time.Sleep(time.Second)
			}
		}
	}()

	e := echo.New()
	r := &RenderWrapper{render.New(render.Options{
		Directory:     "templates",
		Layout:        "layout",
		IsDevelopment: true,
	})}
	e.Renderer = r

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "main", nil)
	})

	e.GET("/ws", func(c echo.Context) error {
		return m.HandleRequest(c.Response().Writer, c.Request())
	})

	m.HandleConnect(func(mel *melody.Session) {
		// var servers []Server
		// db.Select(&servers, "select * from servers")
		// for _, server := range servers {
		// buff, _ := json.Marshal(server)
		// mel.Write(buff)
		// }
	})

	e.Logger.Fatal(e.Start(":1323"))

	// fmt.Println(SpeedTest("10.14.13.146", "download", *window_size, *test_size))
	// fmt.Println(SpeedTest("10.14.13.146", "upload", *window_size, *test_size))

	// fmt.Println(SpeedTest("10.14.0.10", "download", *window_size, *test_size))
	// fmt.Println(SpeedTest("10.14.0.10", "upload", *window_size, *test_size))
}

func SpeedTest(ip, mode string, window_size, test_size int) float64 {
	block_size := window_size * 1024

	// fmt.Printf("%s %s ", ip, mode)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:1337", ip), 3*time.Second)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	defer conn.Close()

	conn.Write([]byte(fmt.Sprintf("%s|%d|%d", mode, test_size, window_size)))

	buf := make([]byte, block_size)

	t := time.Now()

	switch mode {
	case "download":
		for {
			_, err := conn.Read(buf)

			if err == io.EOF {
				break
			}
		}
	case "upload":
		rand.Read(buf)
		for i := 0; i < test_size*1024*1024/block_size; i++ {
			conn.Write(buf)
		}
	}

	return float64(test_size*8) / time.Since(t).Seconds()
}

type RenderWrapper struct { // We need to wrap the renderer because we need a different signature for echo.
	rnd *render.Render
}

func (r *RenderWrapper) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.rnd.HTML(w, 0, name, data) // The zero status code is overwritten by echo.
}
