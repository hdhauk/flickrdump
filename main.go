package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var username string
var key string
var routines = 4
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)
var outputFolder string
var wantAlbumPhotos = true
var wantNonAlbumPhotos = true

func main() {
	// Parse command line arguments.
	flag.StringVar(&username, "u", "", "username from which the dump is happening")
	flag.StringVar(&outputFolder, "o", "", "output folder, if not set default to current directory")
	flag.IntVar(&routines, "n", routines, "number of concurrent downloads")
	flag.BoolVar(&wantAlbumPhotos, "albums", wantAlbumPhotos, "download photos in albums")
	flag.BoolVar(&wantNonAlbumPhotos, "unsorted", wantNonAlbumPhotos, "download photos outside of albums")
	flag.Parse()

	if username == "" {
		username, routines = getUserArgs()
	}

	userID, err1 := getUserIDByUsername(username, key)
	if err1 != nil {
		mainlogger.Fatalf("unable to get user id for user %s: %s\n", username, err1.Error())
	}

	photosDownloaded := []Photo{} // Keep track of photos that are downloaded
	albums, err2 := getAlbumsByUser(userID, key)
	if err2 != nil {
		mainlogger.Fatalf("unable to get albums : %s\n", err2.Error())
	}

	if outputFolder == "" {
		outputFolder, _ = os.Getwd()
	}

	// Download albums into folders for each album.
	for _, album := range albums {
		fmt.Printf("Inspecting %s\n", album.Title)
		// Fetch photos in each album.
		albumPhotos, err := getPhotosInAlbum(album, userID, key)
		if err != nil {
			mainlogger.Printf("Failed to get photos in album %s: %s", album.Title.Content, err.Error())
		}
		photosDownloaded = append(photosDownloaded, albumPhotos...)

		// Download album
		if wantAlbumPhotos {
			albumFolder := createFolder(outputFolder, sanitize(album.Title.Content))
			downloadPhotosAndReport(albumPhotos, albumFolder, key)
		}
	}

	if wantNonAlbumPhotos {
		// Download all photos NOT in an album
		allPhotos, err3 := getAllUserPhotos(userID, key)
		if err3 != nil {
			mainlogger.Fatalln(err3)
		}
		remaining := notAlreadyDownloaded(photosDownloaded, allPhotos)
		path := createFolder(outputFolder, "Misc")
		downloadPhotosAndReport(remaining, path, key)
	}

}
