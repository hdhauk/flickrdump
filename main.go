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

	userID, err := getUserIDByUsername(username)
	if err != nil {
		mainlogger.Fatalf("unable to get user id for user %s: %s\n", username, err.Error())
	}

	albums, err := getAlbumsByUser(userID)
	if err != nil {
		mainlogger.Fatalf("unable to get albums : %s\n", err.Error())
	}

	for _, album := range albums {
		fmt.Printf("Downloading %s\n", album.Title)
		downloadAlbum(album, userID)
	}

}

// PhotoSetResp is the full response received from the photoset.getPhotos endpoint.
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
	// Compose photolist request
	modifiedUID := strings.Replace(userID, "@", "%40", -1)
	lstReq := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=%s&photoset_id=%s&user_id=%s&format=json&nojsoncallback=1", key, a.ID, modifiedUID)

	// Fetch list of all photos in album.
	lstResp, err := http.Get(lstReq)
	if err != nil {
		return fmt.Errorf("failed to fetch photo list from flickr.photosets.getPhotos: %s", err.Error())
	}
	defer lstResp.Body.Close()

	// Decode photolist
	var psr PhotoSetResp
	err = json.NewDecoder(lstResp.Body).Decode(&psr)
	if err != nil {
		return fmt.Errorf("unable to decode response from flickr.photosets.getPhotos: %s", err.Error())
	}
	if psr.Status != "ok" {
		return fmt.Errorf("response returned non-ok status: %s", psr.Status)
	}

	// Create a valid filepath, and create new folder if it doesn't exist.
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	folderName := sanitize(a.Title.Content)
	filePath := fmt.Sprintf("%s/%s", workingDir, folderName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Mkdir(filePath, 0700)
	}

	// Set up communication
	total := len(psr.Set.Photos) // number of photos in album
	var wg sync.WaitGroup
	wg.Add(total)
	progressCh := make(chan int)          // communicate to progress monitor fetch succeeded
	skippedCh := make(chan string, total) // communicate to progress monitor fetch skipped
	errorCh := make(chan error, total)    // communicate to progress monitor fetch failed
	printSummaryCh := make(chan int)      // communicate to progress monitor to print summary
	summaryDoneCh := make(chan int)       // communicate from progress monitor that summary is done and goroutine ends
	sem := make(chan int, routines)       // semaphore limiting number of workers

	// Spawn progress monitor
	go func(numTotal int) {
		numDone := 0
		for {
			select {
			case <-progressCh:
				numDone++
				fmt.Printf("\r--> Processing: %d/%d", numDone, numTotal)
			case <-printSummaryCh:
				fmt.Printf("\n--> Done: %d/%d (%d skipped)\tFailed:%d\n", numDone, numTotal, len(skippedCh), len(errorCh))
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
			skipped, err := downloadAndSavePhoto(url, fp, fn)
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

	// Wait for all workers to complete
	wg.Wait()
	close(errorCh)

	// Print summary before quitting
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

func downloadAndSavePhoto(url, filePath, fileName string) (skipped bool, err error) {
	// Create safe filename with correct suffix.
	cleanFileName := sanitize(fileName)
	fileSuffix := path.Ext(url)
	path := fmt.Sprintf("%s/%s%s", filePath, cleanFileName, fileSuffix)

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

func sanitize(dirty string) string {
	illegalSubStrings := []struct {
		Illegal     string
		ReplaceWith string
	}{
		{"/", "-"},
		{"<", "-"},
		{">", "-"},
		{":", "-"},
		{"\\", "-"},
		{"|", "-"},
		{"?", "-"},
		{"*", "-"},
	}

	clean := dirty
	for _, v := range illegalSubStrings {
		clean = strings.Replace(clean, v.Illegal, v.ReplaceWith, -1)
	}
	return clean

}
