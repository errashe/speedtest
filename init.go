package main

import (
	. "fmt"
	"log"
	"net"
	"os"
	"time"

	. "./structs"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/urfave/cli"
)

var db *storm.DB
var err error

func main() {
	// defer db.Close()
	app := cli.NewApp()
	app.Version = "0.1.4"
	app.Author = "Александр errashe Бутенко"
	app.Usage = "утилита для управления хостами в программе мониторинга"

	app.HideHelp = true
	// app.HideVersion = true

	app.Before = func(c *cli.Context) error {
		db, err = storm.Open("my.db")
		return err
	}

	app.After = func(c *cli.Context) error {
		return db.Close()
	}

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "[host] для добавления в базу",
			Action: func(c *cli.Context) error {
				host := c.Args().First()
				_, err := net.LookupHost(host)
				if err != nil {
					return err
				}

				s := new(Server)
				s.IP = host
				s.Timestamp = time.Now()

				if err = db.Save(s); err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "[host] для удаления из базы",
			Action: func(c *cli.Context) error {
				host := c.Args().First()
				_, err := net.LookupHost(host)
				if err != nil {
					return err
				}
				var server Server
				if err = db.Select(q.Eq("IP", host)).First(&server); err != nil {
					return err
				}
				return db.DeleteStruct(&server)
			},
		},
		{
			Name:    "removeAll",
			Aliases: []string{"ra"},
			Usage:   "Очистка всей базы",
			Action: func(c *cli.Context) error {
				err := db.Select().Delete(new(Server))
				if err == storm.ErrNotFound {
					log.Println(err.Error())
				} else if err != nil {
					return err
				}

				err = db.Select().Delete(new(History))
				if err == storm.ErrNotFound {
					log.Println(err.Error())
				} else if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "removeHistory",
			Aliases: []string{"rh"},
			Usage:   "Очистка истории",
			Action: func(c *cli.Context) error {
				err = db.Select().Delete(new(History))
				if err == storm.ErrNotFound {
					log.Println(err.Error())
				} else if err != nil {
					return err
				}

				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "список всех хостов",
			Action: func(c *cli.Context) error {
				var servers []Server

				if err = db.Select().OrderBy("Timestamp").Find(&servers); err != nil {
					return err
				}

				for _, server := range servers {
					Println(server)
				}

				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
