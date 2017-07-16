package main

import (
	"fmt"
	"log"
	"os"
)

type Option struct {
	Username       string
	UserURL        string
	OutputFolder   string
	IgnoreUnsorted bool
	IgnoreAlbums   bool
	NumRoutines    int
	NoDownload     bool
}

var key string
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)

func main() {
	opts := parseArgs()

	var userID string
	if opts.Username == "" && opts.UserURL == "" {
		os.Exit(0)
	} else if opts.Username != "" {
		var err error
		userID, err = getUserIDByUsername(opts.Username, key)
		if err != nil {
			mainlogger.Fatalf("unable to get user-id from username %s: %s\n", opts.Username, err.Error())
		}
	} else if opts.UserURL != "" {
		var err error
		if userID, err = getUserIDByURL(opts.UserURL, key); err != nil {
			mainlogger.Fatalln(err)
		}
	}

	albums, err := getAlbumsByUserID(userID, key)
	if err != nil {
		mainlogger.Fatalln(err)
	}

	if opts.OutputFolder == "" {
		opts.OutputFolder, _ = os.Getwd()
	}

	// Download albums into folders for each album.
	photosDownloaded := []Photo{} // Keep track of photos that are downloaded
	for _, album := range albums {
		fmt.Printf("Inspecting %s\n", album.Title)
		// Fetch photos in each album.
		albumPhotos, err := getPhotosInAlbum(album, userID, key)
		if err != nil {
			mainlogger.Printf("Failed to get photos in album %s: %s", album.Title.Content, err.Error())
		}
		photosDownloaded = append(photosDownloaded, albumPhotos...)

		if opts.NoDownload {
			fmt.Printf("--> %d photos\n", len(albumPhotos))
		} else if opts.IgnoreAlbums == false {
			albumFolder := createFolder(opts.OutputFolder, sanitize(album.Title.Content))
			downloadPhotosAndReport(albumPhotos, albumFolder, key, opts.NumRoutines)
		}
	}

	if opts.IgnoreUnsorted == false {
		fmt.Println("Inspecting photos not associated with any album")
		// Download all photos NOT in an album
		allPhotos, err := getAllUserPhotos(userID, key)
		if err != nil {
			mainlogger.Fatalln(err)
		}
		remaining := notAlreadyDownloaded(photosDownloaded, allPhotos)
		path := createFolder(opts.OutputFolder, "Misc")

		if opts.NoDownload {
			fmt.Printf("--> %d photos found\n", len(remaining))

		} else if !opts.IgnoreUnsorted {
			downloadPhotosAndReport(remaining, path, key, opts.NumRoutines)
		}
	}
}
