package main

import (
	"os"

	"github.com/urfave/cli"
)

func parseArgs(args []string) Option {
	// Parse command line arguments.
	opts := Option{}
	app := cli.NewApp()
	app.Name = "flickrdump"
	app.Usage = "Download photos from Flickr, the fast way!"
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:  "username",
			Usage: "The username to download from",
			Action: func(c *cli.Context) error {
				opts.Username = c.Args().First()
				return nil
			},
		},
		{
			Name:  "url",
			Usage: "The url of the user to download from",
			Action: func(c *cli.Context) error {
				opts.UserURL = c.Args().First()
				return nil
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "ignore-albums",
			Usage:       "Do not download photos that are sorted into albums",
			Destination: &opts.IgnoreAlbums,
		},
		cli.BoolFlag{
			Name:        "ignore-unsorted",
			Usage:       "Do not download photos that aren't part of any albums",
			Destination: &opts.IgnoreUnsorted,
		},
		cli.IntFlag{
			Name:        "threads",
			Usage:       "Number of concurrent downloads",
			Value:       4,
			Destination: &opts.NumRoutines,
		},
		cli.BoolFlag{
			Name:        "no-download",
			Usage:       "Only scan user, no files will be downloaded.",
			Destination: &opts.NoDownload,
		},
	}

	app.Run(os.Args)
	return opts
}
