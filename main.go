package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/kennygrant/sanitize"
)

type Content struct {
	Content string `json:"_content"`
}

var username = ""
var key = ""
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)

func main() {
	flag.StringVar(&key, "key", "", "API Key")
	flag.StringVar(&username, "u", "", "username from which the dump is happening")
	flag.Parse()

	uid, e := getUserIDByUsername(username)
	if e != nil {
		mainlogger.Fatalf("unable to get user id for user %s: %s\n", username, e.Error())
	}
	albums, e := getAlbumsByUser(uid)
	if e != nil {
		mainlogger.Fatalf("unable to get albums : %s\n", e.Error())
	}

	for _, a := range albums {
		fmt.Println("============================================================================")
		fmt.Printf("    Downloading %s\n", a.Title)
		fmt.Println("----------------------------------------------------------------------------")
		downloadAlbum(a, uid)
	}

}

type PhotoSetResp struct {
	Set    PhotoSet `json:"photoset"`
	Status string   `json:"stat"`
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

func downloadAlbum(a Album, userID string) error {
	userID = strings.Replace(userID, "@", "%40", -1)
	lstReq := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=%s&photoset_id=%s&user_id=%s&format=json&nojsoncallback=1", key, a.ID, userID)

	lstResp, err := http.Get(lstReq)
	if err != nil {
		panic(err)
	}
	var psr PhotoSetResp
	err = json.NewDecoder(lstResp.Body).Decode(&psr)
	if err != nil {
		return err
	}
	if psr.Status != "ok" {
		return fmt.Errorf("unable to get albums photos")
	}

	// Create filePath
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(dir)
		mainlogger.Fatal(err)
	}
	folderName := strings.Replace(a.Title.Content, "/", "_", -1)
	filePath := fmt.Sprintf("%s/%s", dir, folderName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Mkdir(filePath, 0700)
	}

	var wg sync.WaitGroup
	for _, p := range psr.Set.Photos {
		wg.Add(1)
		go func(photoID, fp, fn string) {
			defer wg.Done()
			url := getDownloadLink(photoID)
			downloadAndSavePhoto(url, fp, fn)
		}(p.ID, filePath, p.Title)

	}
	wg.Wait()
	return nil
}

func getDownloadLink(photoID string) string {
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photos.getSizes&api_key=%s&photo_id=%s&format=json&nojsoncallback=1", key, photoID)
	resp, err := http.Get(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var sr SizesResp
	json.NewDecoder(resp.Body).Decode(&sr)
	for _, s := range sr.Sizes.Versions {
		if s.Label == "Original" {
			return s.Src
		}
	}
	return ""
}

func downloadAndSavePhoto(url, filePath, fileName string) {
	fileSuffix := path.Ext(url)
	resp, e := http.Get(url)
	if e != nil {
		mainlogger.Printf("Cannot download photo %s: %s\n", fileName, e.Error())
		return
	}
	defer resp.Body.Close()

	//open a file for writing
	cleanFileName := sanitize.Path(fileName)
	file, err := os.Create(fmt.Sprintf("%s/%s%s", filePath, cleanFileName, fileSuffix))
	if err != nil {
		mainlogger.Println(err)
		return
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		mainlogger.Println(err)
		return
	}
	file.Close()
	fmt.Printf("Download complete: %s\n", fileName)

}
