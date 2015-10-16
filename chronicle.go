package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "chronicle"
	app.Version = "0.0.1"
	app.Usage = "requirement management done right"
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
			Name:  "commit",
			Usage: "complete a task on the list",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
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
