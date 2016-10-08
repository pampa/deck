package main

import (
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Usage = "hands off package manager"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config file to use instead of /etc/deckrc or $HOME/.deckrc",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "print debug info to stderr",
		},
	}

	app.Before = func(c *cli.Context) error {
		log.Verbose = c.GlobalBool("debug")
		if cFile, err := getConfigFile(c.GlobalString("config")); err == nil {
			deck.Init(cFile)
		} else {
			log.Error(err)
		}
		return nil
	}

	app.After = func(c *cli.Context) error {
		log.Debug("closing database")
		deck.Close()
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:    "scan",
			Aliases: []string{"s"},
			Usage:   "scan the filesystem for changes",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "hash, s",
					Usage: "use sha1 to compare files",
				},
				cli.BoolFlag{
					Name:  "pick, p",
					Usage: "pick new files",
				},
			},
			Action: func(c *cli.Context) {
				deck.Scan(c.Bool("hash"), c.Bool("pick"))
			},
		},
		{
			Name:    "pick",
			Aliases: []string{"p"},
			Usage:   "pick file for further processing",
			Action: func(c *cli.Context) {
				deck.Pick(c.Args())
			},
		},
		{
			Name:    "unpick",
			Aliases: []string{"u"},
			Usage:   "unpick file",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "unpick all files",
				},
			},
			Action: func(c *cli.Context) {
				deck.Unpick(c.Bool("all"), c.Args())
			},
		},
		{
			Name:  "commit",
			Usage: "commit picked files to index, adding package and version tags",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "package, p",
					Usage: "package name",
				},
				cli.StringFlag{
					Name:  "version, v",
					Usage: "package version",
				},
			},
			Action: func(c *cli.Context) {
				pak := c.String("package")
				ver := c.String("version")

				if pak == "" && ver == "" {
					log.Error("--package and --version flags are required")
				} else if pak == "" {
					log.Error("--package flag is required")
				} else if ver == "" {
					log.Error("--version flag is required")
				} else {
					deck.Commit(pak, ver)
				}
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list all packages in index",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "version, v",
					Usage: "do not print version number",
				},
			},
			Action: func(c *cli.Context) {
				deck.List(c.Bool("version"))
			},
		},
		{
			Name:    "show",
			Aliases: []string{"o"},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all",
					Usage: "show all tracked files",
				},
			},
			Usage: "show files in package",
			Action: func(c *cli.Context) {
				if c.Bool("all") && len(c.Args()) > 0 {
					log.Error("cant use --all with package name")
				} else if c.Bool("all") {
					deck.Show("", true)
				} else if len(c.Args()) == 0 {
					log.Error("show what?")
				} else {
					deck.Show(c.Args()[0], false)
				}
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "remove file from index",
			Action: func(c *cli.Context) {
				deck.Remove(c.Args())
			},
		},
		{
			Name:  "reset",
			Usage: "reset file to its previous state",
			Action: func(c *cli.Context) {
				deck.Reset(c.Args())
			},
		},
		{
			Name:  "uninstall",
			Usage: "uninstall package",
			Action: func(c *cli.Context) {
				deck.Uninstall(c.Args()[0])
			},
		},
		{
			Name:    "which",
			Aliases: []string{"w", "who", "what"},
			Usage:   "show which package a file belongs to",
			Action: func(c *cli.Context) {
				deck.Which(c.Args())
			},
		},
		{
			Name:    "doctor",
			Aliases: []string{"doc", "d"},
			Usage:   "run database sanity checks",
			Action: func(c *cli.Context) {
				deck.Doctor()
			},
		},
	}
	app.Run(os.Args)
}
