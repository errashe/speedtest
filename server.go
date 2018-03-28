package main

import (
	// "bytes"
	"encoding/json"
	"flag"
	. "fmt"
	"io"

	. "./structs"

	"github.com/asdine/storm"
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

	db, err = storm.Open("my.db")
	if err != nil {
		Println(err)
	}
	defer db.Close()

	db.Init(&Server{})

	// switch err = db.Save(&Server{0, "localhost", 0, 0, 0, time.Now()}); err {
	// case storm.ErrAlreadyExists:
	// 	break
	// default:
	// 	Println(err)
	// }

	exitSaver()

	m = melody.New()

	go speedTester(*window_size, *test_size)

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
		if err = db.Select().OrderBy("ID").Find(&servers); err != nil {
			Println(err)
		}
		for _, server := range servers {
			buff, _ := json.Marshal(server)
			mel.Write(buff)
		}
	})

	e.Logger.Fatal(e.Start(":1323"))
}

type RenderWrapper struct { // We need to wrap the renderer because we need a different signature for echo.
	rnd *render.Render
}

func (r *RenderWrapper) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.rnd.HTML(w, 0, name, data) // The zero status code is overwritten by echo.
}
