package main

import (
	"flag"
	. "fmt"
	"time"

	"github.com/asdine/storm"
)

func main() {
	host := flag.String("host", "localhost", "hostname for adding to database")
	flag.Parse()

	db, err := storm.Open("my.db")
	if err != nil {
		Println(err)
	}
	defer db.Close()

	db.Init(&Server{})

	switch err = db.Save(&Server{0, *host, 0, 0, 0, time.Now()}); err {
	case storm.ErrAlreadyExists:
		Println("Already exists")
	case nil:
		Println("Writed")
	default:
		Println(err)
	}

}
