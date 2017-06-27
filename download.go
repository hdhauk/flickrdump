package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
)

// downloadPhotosAndReport downloads the given photos into the destination folder
func downloadPhotosAndReport(photos []Photo, dstPath, APIkey string) {
	// Set up communication
	total := len(photos) // number of photos in album
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
	for index, p := range photos {
		prefix := strconv.Itoa(index + 1)
		go func(photoID, fp, fn string) {
			defer wg.Done()
			sem <- 1
			url, err := getPhotoDownloadLink(photoID, APIkey)
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
		}(p.ID, dstPath, prefix+"-"+p.Title)
	}

	// Wait for all workers to complete
	wg.Wait()
	close(errorCh)

	// Print summary before quitting
	printSummaryCh <- 1
	<-summaryDoneCh
}

func getPhotoDownloadLink(photoID, APIkey string) (string, error) {
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photos.getSizes&api_key=%s&photo_id=%s&format=json&nojsoncallback=1", APIkey, photoID)
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
