package main

import (
	"os"
	"path/filepath"

	"github.com/Benefactory/chronicle/database"
	"github.com/Benefactory/chronicle/walker"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "chronicle"
	app.Version = "0.0.1"
	app.Usage = "requirement management done right"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path",
			Value: ".",
			Usage: "Path to git repo. Default is current working directory",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "add",
			Usage: "add template requirment to specific a file",
			Subcommands: []cli.Command{
				{
					Name:  "feature",
					Usage: "add a new feature",
					Action: func(c *cli.Context) {
						println("Filename: ", c.Args().First())
						println("Req title: ", c.Args()[1])
					},
				},
				{
					Name:  "goal",
					Usage: "add a new goal",
					Action: func(c *cli.Context) {
						println("Filename: ", c.Args().First())
						println("Req title: ", c.Args()[1])
					},
				},
			},
		},
		{
			Name:  "test",
			Usage: "run test code and requirments",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
		{
			Name:  "commit",
			Usage: "generate an chronicle report",
			Action: func(c *cli.Context) {
				var db *database.Database
				if len(c.String("path")) > 0 {
					setupChronicle(c.String("path"))
					db = database.NewDatabase("path")
					walker.UpdateRepo(c.String("path"), db)
				} else {
					setupChronicle("")
					db = database.NewDatabase("")
					walker.UpdateRepo("", db)
				}
				db.Close()
			},
		},
		{
			Name:  "server",
			Usage: "start the webserver",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
		{
			Name:  "gui",
			Usage: "start the desktop application",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
	}
	app.Run(os.Args)
}

func setupChronicle(rootPath string) error {
	// TODO: Check for git-root repo

	os.Mkdir("."+string(filepath.Separator)+rootPath+".git"+string(filepath.Separator)+"chronicle", 0777)

	return nil
}
