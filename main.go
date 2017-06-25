package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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

func main() {
	// Parse command line arguments.
	flag.StringVar(&key, "key", "", "API Key")
	flag.StringVar(&username, "u", "", "username from which the dump is happening")
	flag.IntVar(&routines, "n", routines, "number of concurrent downloads")
	flag.Parse()

	if username == "" || key == "" {
		username, key, routines = getUserArgs()
	}

	userID, err := getUserIDByUsername(username, key)
	if err != nil {
		mainlogger.Fatalf("unable to get user id for user %s: %s\n", username, err.Error())
	}

	albums, err := getAlbumsByUser(userID, key)
	if err != nil {
		mainlogger.Fatalf("unable to get albums : %s\n", err.Error())
	}

	photosDownloaded := []Photo{}

	for _, album := range albums {
		fmt.Printf("Downloading %s\n", album.Title)
		photosInAlbum, _ := downloadAlbum(album, userID, key)
		photosDownloaded = append(photosDownloaded, photosInAlbum...)
	}

	// Download all photos NOT in an album
	fmt.Println("Downloading unsorted photos")
	allPhotos, e := getAllUserPhotos(userID, key)
	if e != nil {
		fmt.Println(e)
	}
	remaining := notAlreadyDownloaded(photosDownloaded, allPhotos)
	workingDir, _ := os.Getwd()
	path := createFolder(workingDir, "unsorted")
	downloadPhotosAndReport(remaining, path, key)

}

// PhotoSetResp is the full response received from the photoset.getPhotos endpoint.
type PhotoSetResp struct {
	Set    PhotoSet `json:"photoset"`
	Status string   `json:"stat"`
}

// returns all downloaded all photos in the album.
func downloadAlbum(a Album, userID, APIkey string) ([]Photo, error) {
	// Compose photolist request
	modifiedUID := strings.Replace(userID, "@", "%40", -1)
	lstReq := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=%s&photoset_id=%s&user_id=%s&format=json&nojsoncallback=1", APIkey, a.ID, modifiedUID)

	// Fetch list of all photos in album.
	lstResp, err := http.Get(lstReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch photo list from flickr.photosets.getPhotos: %s", err.Error())
	}
	defer lstResp.Body.Close()

	// Decode photolist
	var psr PhotoSetResp
	err = json.NewDecoder(lstResp.Body).Decode(&psr)
	if err != nil {
		return nil, fmt.Errorf("unable to decode response from flickr.photosets.getPhotos: %s", err.Error())
	}
	if psr.Status != "ok" {
		return nil, fmt.Errorf("response returned non-ok status: %s", psr.Status)
	}

	// Create a valid filepath, and create new folder if it doesn't exist.
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	folderName := sanitize(a.Title.Content)
	filePath := path.Join(workingDir, folderName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Mkdir(filePath, 0700)
	}

	downloadPhotosAndReport(psr.Set.Photos, filePath, APIkey)

	return psr.Set.Photos, nil
}

func downloadAndSavePhoto(url, filePath, fileName string) (skipped bool, err error) {
	// Create safe filename with correct suffix.
	cleanFileName := sanitize(fileName)
	fileSuffix := path.Ext(url)
	path := path.Join(filePath, cleanFileName) + fileSuffix

	// Skip if file already exist.
	if _, err := os.Stat(path); err == nil {
		return true, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return false, err
	}

	// Download photo
	resp, e := http.Get(url)
	if e != nil {
		file.Close()
		os.Remove(path)
		return false, e
	}
	defer resp.Body.Close()

	// Copy photo from response to file.
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		os.Remove(path)
		return false, err
	}
	file.Close()
	return false, nil
}

func getUserArgs() (user, key string, workers int) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Flickr API key: ")
	key, _ = reader.ReadString('\n')
	key = strings.TrimSuffix(key, "\n")

	fmt.Print("Flickr user to download all albums from: ")
	user, _ = reader.ReadString('\n')
	user = strings.TrimSuffix(user, "\n")

	fmt.Print("Number of parallell downloads: ")
	fmt.Scanf("%d", &workers)
	fmt.Println()

	return
}

func getAllUserPhotos(userID, APIkey string) ([]Photo, error) {
	// Compose API URL
	modifiedID := strings.Replace(userID, "@", "%40", -1)
	url := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.people.getPhotos&api_key=%s&user_id=%s&format=json&nojsoncallback=1", APIkey, modifiedID)

	// Fetch JSON from Flickr API
	resp, err := http.Get(url)
	if err != nil {
		mainlogger.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Decode JSON
	target := struct {
		Photos struct {
			Photos []Photo `json:"photo"`
		} `json:"photos"`
		Status string `json:"stat"`
	}{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return nil, jsonErr
	}
	if target.Status != "ok" {
		return nil, fmt.Errorf("unable to get all photos")
	}
	return target.Photos.Photos, nil

}

type PhotoSet struct {
	Title  string  `json:"title"`
	Photos []Photo `json:"photo"`
}
type Photo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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
