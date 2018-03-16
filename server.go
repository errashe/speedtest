package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/olahol/melody.v1"
	"gopkg.in/unrolled/render.v1"
)

var schema = `
	CREATE TABLE IF NOT EXISTS 'servers' (
		id integer primary key autoincrement,
		ip varchar(15) unique,
		download double default 0,
		upload double default 0,
		timestamp timestamp default current_timestamp
	)
`

var db *sqlx.DB

type Server struct {
	ID        int       `db:"id"`
	IP        string    `db:"ip"`
	Download  float64   `db:"download"`
	Upload    float64   `db:"upload"`
	Timestamp time.Time `db:"timestamp"`
}

func (s Server) String() string {
	return fmt.Sprintf("%d, %s, %0.2f, %0.2f, %s", s.ID, s.IP, s.Download, s.Upload, s.Timestamp)
}

func main() {
	window_size := flag.Int("w", 64, "size of window, which one gonna be send")
	test_size := flag.Int("t", 40, "megabytes of data, which one gonna be send")

	flag.Parse()

	m := melody.New()

	db, err := sqlx.Connect("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.MustExec(schema)
	// db.MustExec("INSERT INTO `servers` ('ip') VALUES('10.14.0.10')")
	// db.MustExec("INSERT INTO `servers` ('ip') VALUES('10.14.13.146')")

	go func() {
		for {
			var tmp Server
			err = db.Get(&tmp, "select * from `servers` order by timestamp limit 1")
			if err != nil {
				fmt.Println(err)
				time.Sleep(3 * time.Second)
				continue
			}

			tmp.Download = SpeedTest(tmp.IP, "download", *window_size, *test_size)
			tmp.Upload = SpeedTest(tmp.IP, "upload", *window_size, *test_size)
			tmp.Timestamp = time.Now()

			buff, _ := json.Marshal(tmp)
			m.Broadcast(buff)

			db.MustExec("UPDATE `servers` SET `ip` = ?, `download` = ?, `upload` = ?, `timestamp` = ? WHERE `id` = ?",
				tmp.IP, tmp.Download, tmp.Upload, tmp.Timestamp, tmp.ID)

			fmt.Println(tmp)

			time.Sleep(time.Second)
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
		var servers []Server
		db.Select(&servers, "select * from servers")
		for _, server := range servers {
			buff, _ := json.Marshal(server)
			mel.Write(buff)
		}
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
