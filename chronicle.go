package main

import (
	"os"

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
				if len(c.String("path")) > 0 {
					walker.UpdateRepo(c.String("path"))
				} else {
					walker.UpdateRepo(".")
				}

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
