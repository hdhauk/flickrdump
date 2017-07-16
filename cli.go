package main

import (
	"flag"
	"fmt"
	"os"
)

func parseArgs() Option {
	var opts Option
	flag.StringVar(&opts.Username, "user", "", "Username for account to dump from. Note that this is not always what appares on the profile page.")
	flag.StringVar(&opts.UserURL, "url", "", "URL to the profile page of the account to dump from.")
	flag.BoolVar(&opts.IgnoreUnsorted, "nounsorted", false, "Ignore photos not found in any albums.")
	flag.BoolVar(&opts.IgnoreAlbums, "noalbums", false, "Ignore photos that are in one or more albums.")
	flag.IntVar(&opts.NumRoutines, "n", 4, "Number of concurrent downloads.")

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Unable to get working directory: %s. \n Using root folder (/) instead.", err.Error())
		workingDir = "/"
	}
	flag.StringVar(&opts.OutputFolder, "out", workingDir, "Destination folder.")
	flag.BoolVar(&opts.NoDownload, "nodownload", false, "Only scan photos, skip actual download. Useful to check for number of photos")

	flag.Parse()
	return opts
}
