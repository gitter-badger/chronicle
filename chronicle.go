package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	// Create local database folder
	os.Mkdir("."+string(filepath.Separator)+rootPath+".chronicle", 0777)

	// Check if the local database is ignored
	file, err := os.Open("." + string(filepath.Separator) + rootPath + ".gitignore")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	var b bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		b = strings.EqualFold(scanner.Text(), ".chronicle")
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Update .gitignore file to prevent the local database to be commited.
	if !b {
		f, err := os.OpenFile("."+string(filepath.Separator)+rootPath+".gitignore", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err = f.WriteString("\n# Added by chronicle, ignoring local database \n.chronicle"); err != nil {
			panic(err)
		}
	}
	return err
}
