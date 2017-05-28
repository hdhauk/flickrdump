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
var routines = 4
var mainlogger = log.New(os.Stderr, "[main] ", log.Ltime|log.Lshortfile)

func main() {
	flag.StringVar(&key, "key", "", "API Key")
	flag.StringVar(&username, "u", "", "username from which the dump is happening")
	flag.IntVar(&routines, "n", routines, "number of concurrent downloads")
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
		fmt.Printf("Downloading %s\n", a.Title)
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
	defer lstResp.Body.Close()
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

	// Progress bar
	total := len(psr.Set.Photos)
	var wg sync.WaitGroup
	wg.Add(total)
	progressCh := make(chan int)
	printSummaryCh := make(chan int)
	summaryDoneCh := make(chan int)
	errorCh := make(chan error, total)
	skippedCh := make(chan string, total)
	sem := make(chan int, routines)
	go func(total int) {
		i := 0
		for {
			select {
			case <-progressCh:
				i++
				fmt.Printf("\r--> Processing: %d/%d", i, total)
			case <-printSummaryCh:
				fmt.Printf("\n--> Done: %d/%d (%d skipped)\tFailed:%d\n", i, total, len(skippedCh), len(errorCh))
				for e := range errorCh {
					fmt.Println(e)
				}
				fmt.Println()
				summaryDoneCh <- 1
				return
			}
		}
	}(total)

	// Spawn workers
	for _, p := range psr.Set.Photos {
		go func(photoID, fp, fn string) {
			defer wg.Done()
			sem <- 1
			url, err := getDownloadLink(photoID)
			if err != nil {
				<-sem
				errorCh <- fmt.Errorf("%s > %s", photoID, err)
				return
			}
			err, skipped := downloadAndSavePhoto(url, fp, fn)
			<-sem
			if skipped {
				skippedCh <- fn
				progressCh <- 1
			} else if err != nil {
				errorCh <- fmt.Errorf("%s > %s", fn, err)
			} else {
				progressCh <- 1
			}
		}(p.ID, filePath, p.Title)
	}
	wg.Wait()
	close(errorCh)
	printSummaryCh <- 1
	<-summaryDoneCh

	return nil
}

func getDownloadLink(photoID string) (string, error) {
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photos.getSizes&api_key=%s&photo_id=%s&format=json&nojsoncallback=1", key, photoID)
	resp, err := http.Get(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var sr SizesResp
	json.NewDecoder(resp.Body).Decode(&sr)
	for _, s := range sr.Sizes.Versions {
		if s.Label == "Original" {
			return s.Src, nil
		}
	}
	return "", fmt.Errorf("unable to find original")
}

func downloadAndSavePhoto(url, filePath, fileName string) (err error, skipped bool) {
	//open a file for writing
	cleanFileName := sanitize.Path(fileName)
	fileSuffix := path.Ext(url)
	path := fmt.Sprintf("%s/%s%s", filePath, cleanFileName, fileSuffix)
	if _, err := os.Stat(path); err == nil {
		// File already exist
		return nil, true
	}
	file, err := os.Create(path)
	if err != nil {
		return err, false
	}
	resp, e := http.Get(url)
	if e != nil {
		file.Close()
		os.Remove(path)
		return e, false
	}
	defer resp.Body.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		os.Remove(path)
		return err, false
	}
	file.Close()
	return nil, false
}
