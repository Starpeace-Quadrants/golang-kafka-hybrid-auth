package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sp-auth-server/apiserver"
	"sp-auth-server/storage/mongo"
)

const (
	apiServerAddrFlagName string = "addr"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "run the server",
				Action: func(cCtx *cli.Context) error {
					db, err := mongo.NewDatabase()
					if err != nil {
						panic(err.Error())
					}

					db.Start()

					api, err := apiserver.NewAPIServer(db.Client)
					if err != nil {
						panic(err.Error())
					}

					api.Start()
					return nil
				},
			},
			{
				Name:    "migrations",
				Aliases: []string{"m"},
				Usage:   "run database migrations",
				Subcommands: []*cli.Command{
					{
						Name:  "create database",
						Usage: "create database",
						Action: func(cCtx *cli.Context) error {
							fmt.Println("new task template: ", cCtx.Args().First())
							return nil
						},
					},
					{
						Name:  "migrate",
						Usage: "run migrations",
						Action: func(cCtx *cli.Context) error {
							fmt.Println("removed task template: ", cCtx.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
