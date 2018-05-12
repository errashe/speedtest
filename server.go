package main

import (
	// "bytes"
	"encoding/json"
	"flag"
	. "fmt"
	"os"
	"time"

	. "./funcs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/labstack/echo"
	"gopkg.in/olahol/melody.v1"
	"gopkg.in/unrolled/render.v1"
)

var m *melody.Melody
var db *storm.DB
var err error

func main() {
	window_size := flag.Int("w", 64, "size of window, which one gonna be send")
	test_size := flag.Int("t", 40, "megabytes of data, which one gonna be send")

	flag.Parse()

	if _, err := os.Stat("my.db"); os.IsNotExist(err) {
		Println("Вам нужно инициализировать базу")
		return
	}

	db, err = storm.Open("my.db")
	if err != nil {
		Println(err)
	}
	defer db.Close()

	db.Init(&Server{})
	db.Init(&History{})

	// ExitSaver()

	m = melody.New()

	go SpeedTester(*window_size, *test_size, m, db)

	e := echo.New()
	r := NewRender(render.Options{
		Directory:     "templates",
		Layout:        "layout",
		IsDevelopment: true,
	})
	e.Renderer = r

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "main", nil)
	})

	e.GET("/ws", func(c echo.Context) error {
		return m.HandleRequest(c.Response().Writer, c.Request())
	})

	m.HandleMessage(func(mel *melody.Session, msg []byte) {
		cmd := new(Command)
		err := json.Unmarshal(msg, cmd)
		if err != nil {
			Println(err)
		}

		switch cmd.Command {
		case "up":
			Println(cmd.Value)
		default:
			Println("UNEXPECTED COMMAND FOUND, WTF?!")
		}
	})

	m.HandleConnect(func(mel *melody.Session) {
		var servers []Server
		if err = db.Select().OrderBy("ID").Find(&servers); err != nil {
			Println(err)
		}
		for _, server := range servers {
			buff, _ := json.Marshal(server)
			mel.Write(buff)
		}
	})

	e.GET("/graph", func(e echo.Context) error {
		return e.Render(200, "graph", nil)
	})

	e.GET("/points", func(e echo.Context) error {
		var query Quer
		if err := e.Bind(&query); err != nil {
			return err
		}

		var hs []History
		if err := db.Select(q.Eq("IP", query.IP)).Reverse().Limit(query.Count).Find(&hs); err != nil {
			Println(e.QueryParam("ip"))
			Println(err)
			return err
		}

		var times []time.Time
		var YS1 []float64
		var YS2 []float64

		for _, h := range hs {
			times = append(times, h.Timestamp)
			YS1 = append(YS1, h.Download)
			YS2 = append(YS2, h.Upload)
		}

		return e.JSON(200, echo.Map{"times": times, "YS1": YS1, "YS2": YS2})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
