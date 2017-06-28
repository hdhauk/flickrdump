package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Content is a common object used by the flickr-API. Usually contain a title,
// description or somthing similar.
type Content struct {
	Content string `json:"_content"`
}

// Photo contain both id and title of a Flickr photo
type Photo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// PhotoSet is an alias for sorting purposes.
type PhotoSet []Photo

func (p PhotoSet) Len() int {
	return len(p)
}
func (p PhotoSet) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PhotoSet) Less(i, j int) bool {
	iNum, _ := strconv.Atoi(string(p[i].ID))
	jNum, _ := strconv.Atoi(string(p[j].ID))
	return iNum < jNum
}

// Album is the object containing information on one album.
type Album struct {
	ID          string `json:"id"`
	Title       Content
	Description Content
}

// User is the actual user payload received.
type User struct {
	ID       string `json:"id"`
	NSID     string `json:"nsid"`
	Username Content
}

// Size describe a size version of a flickr photo.
type Size struct {
	Label string `json:"label"`
	Src   string `json:"source"`
}

func getPhotosInAlbum(album Album, userID, APIkey string) ([]Photo, error) {
	// Compose photolist request
	modifiedUID := strings.Replace(userID, "@", "%40", -1)
	url := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=%s&photoset_id=%s&user_id=%s&format=json&nojsoncallback=1", APIkey, album.ID, modifiedUID)

	// Fetch list of photos from API
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch photo list from flickr.photosets.getPhotos: %s", err.Error())
	}
	defer resp.Body.Close()

	// Decode JSON.
	var photoSetResp struct {
		Set struct {
			Title  string  `json:"title"`
			Photos []Photo `json:"photo"`
		} `json:"photoset"`
		Status string `json:"stat"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&photoSetResp); err != nil {
		return nil, fmt.Errorf("failed to decode json: %s", err.Error())
	}
	if photoSetResp.Status != "ok" {
		return nil, fmt.Errorf("api returned bad response: %s", photoSetResp.Status)
	}

	return photoSetResp.Set.Photos, nil
}

func getAllUserPhotos(userID, APIkey string) ([]Photo, error) {
	// Fetch first page
	allPhotos, numPages, err := fetchPhotosByPage(1, userID, APIkey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch photos: %s", err.Error())
	}

	// Fetch all pages
	failedPages := 0
	if numPages < 1 {
		for page := 2; page <= numPages; page++ {
			photosInPage, _, e := fetchPhotosByPage(page, userID, APIkey)
			if e != nil {
				failedPages++
				continue
			}
			allPhotos = append(allPhotos, photosInPage...)
		}
	}
	if failedPages > 0 {
		return allPhotos, fmt.Errorf("%d pages failed to load", failedPages)
	}

	return allPhotos, nil
}

// fetchPhotosByPage returns photos in page and the total number of pages.
func fetchPhotosByPage(page int, userID, APIkey string) ([]Photo, int, error) {
	// Compose API URL
	modifiedID := strings.Replace(userID, "@", "%40", -1)
	url := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.people.getPhotos&api_key=%s&user_id=%s&format=json&nojsoncallback=1&per_page=500&page=%d", APIkey, modifiedID, page)

	// Fetch JSON from Flickr API
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// Decode JSON
	target := struct {
		Photos struct {
			Photos []Photo `json:"photo"`
			Pages  int     `json:"pages"`
		} `json:"photos"`
		Status string `json:"stat"`
	}{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return nil, 0, jsonErr
	}
	if target.Status != "ok" {
		return nil, 0, fmt.Errorf("unable to get all photos")
	}
	return target.Photos.Photos, target.Photos.Pages, nil
}

func getAlbumsByUser(userID, APIkey string) ([]Album, error) {
	// Compose request URL
	modifiedID := strings.Replace(userID, "@", "%40", -1)
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getList&api_key=%s&user_id=%s&format=json&nojsoncallback=1", APIkey, modifiedID)

	resp, err := http.Get(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	var albumResp struct {
		Albums struct {
			Length int     `json:"total"`
			Sets   []Album `json:"photoset"`
		} `json:"photosets"`
		Status string `json:"stat"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&albumResp); err != nil {
		return nil, err
	}
	if albumResp.Status != "ok" {
		return nil, fmt.Errorf("unable to get albums")
	}
	return albumResp.Albums.Sets, nil
}

func getUserIDByUsername(username, APIkey string) (string, error) {
	// Compose API request
	searchFor := strings.Replace(username, " ", "+", -1)
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.people.findByUsername&api_key=%s&username=%s&format=json&nojsoncallback=1", APIkey, searchFor)

	resp, err := http.Get(req)
	if err != nil {
		mainlogger.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	// Decode response
	var userResp struct {
		User   `json:"user"`
		Status string `json:"stat"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	if err != nil {
		return "", err
	}
	if userResp.Status == "fail" {
		return "", fmt.Errorf("failed to find user")
	}
	return userResp.User.ID, nil
}
