package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// Content is a common object used by the flickr-API. Usually contain a title,
// description or somthing similar.
type Content struct {
	Content string `json:"_content"`
}

var username = ""
var key = ""
var routines = 4
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)
var outputFolder = ""
var onlyAlbums = false

func main() {
	// Parse command line arguments.
	flag.StringVar(&key, "key", "", "API Key")
	flag.StringVar(&username, "u", "", "username from which the dump is happening")
	flag.StringVar(&outputFolder, "o", "", "output folder, if not set default to current directory")
	flag.IntVar(&routines, "n", routines, "number of concurrent downloads")
	flag.BoolVar(&onlyAlbums, "onlyalbums", onlyAlbums, "do not download photos that aren't in an album")
	flag.Parse()

	if username == "" || key == "" {
		username, key, routines = getUserArgs()
	}

	userID, err1 := getUserIDByUsername(username, key)
	if err1 != nil {
		mainlogger.Fatalf("unable to get user id for user %s: %s\n", username, err1.Error())
	}

	albums, err2 := getAlbumsByUser(userID, key)
	if err2 != nil {
		mainlogger.Fatalf("unable to get albums : %s\n", err2.Error())
	}

	photosDownloaded := []Photo{}
	if outputFolder == "" {
		outputFolder, _ = os.Getwd()
	}

	// Download albums into folders for each album.
	for _, album := range albums {
		fmt.Printf("Downloading %s\n", album.Title)
		albumFolder := createFolder(outputFolder, sanitize(album.Title.Content))
		albumPhotos, err := getPhotosInAlbum(album, userID, key)
		if err != nil {
			mainlogger.Printf("Failed to get photos in album %s: %s", album.Title.Content, err.Error())
		}
		downloadPhotosAndReport(albumPhotos, albumFolder, key)
		photosDownloaded = append(photosDownloaded, albumPhotos...)
	}

	if onlyAlbums {
		return
	}
	// Download all photos NOT in an album
	fmt.Println("Downloading unsorted photos")
	allPhotos, err3 := getAllUserPhotos(userID, key)
	if err3 != nil {
		mainlogger.Fatalln(err3)
	}
	remaining := notAlreadyDownloaded(photosDownloaded, allPhotos)
	path := createFolder(outputFolder, "unsorted")
	downloadPhotosAndReport(remaining, path, key)

}

type SizesResp struct {
	Sizes  Sizes  `json:"sizes"`
	Status string `json:"stat"`
}
type Sizes struct {
	Versions []Size `json:"size"`
}
type Size struct {
	Label string `json:"label"`
	Src   string `json:"source"`
}
