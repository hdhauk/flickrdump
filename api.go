package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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

const (
	baseURL             = "https://api.flickr.com/services/rest/"
	commonOptions       = "format=json&nojsoncallback=1"
	userIDFromURL       = "flickr.urls.lookupUser"
	userIDFromUsername  = "flickr.people.findByUsername"
	getPhotosFromAlbum  = "flickr.photosets.getPhotos"
	getPhotosFromUser   = "flickr.people.getPhotos"
	getAlbumsFromUserID = "flickr.photosets.getList"
)

func getUserIDByURL(url, APIKey string) (username string, err error) {
	// Clean up url
	cleanURL := strings.Replace(url, "https://www.", "", 1)
	cleanURL = strings.Replace(cleanURL, "/", "%2F", -1)

	// Contact API
	apiURL := fmt.Sprintf("%s?method=%s&url=%s&api_key=%s&%s", baseURL, userIDFromURL, cleanURL, APIKey, commonOptions)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", errors.Wrap(err, "GET request to "+userIDFromURL+" failed")
	}
	defer resp.Body.Close()

	// Decode JSON
	var userSearch struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		Status string `json:"stat"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userSearch); err != nil {
		return "", errors.Wrap(err, "json decoding failed")
	} else if userSearch.Status != "ok" {
		return "", fmt.Errorf("failed to find user, api returned: %s", userSearch.Status)
	}
	return userSearch.User.ID, nil

}

func getUserIDByUsername(username, APIKey string) (string, error) {
	// Compose API request
	searchFor := strings.Replace(username, " ", "+", -1)
	apiURL := fmt.Sprintf("%s?method=%s&api_key=%s&username=%s&%s", baseURL, userIDFromUsername, APIKey, searchFor, commonOptions)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", errors.Wrapf(err, "GET request to %s failed", userIDFromUsername)
	}
	defer resp.Body.Close()

	// Decode response
	var userResp struct {
		User   `json:"user"`
		Status string `json:"stat"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	if err != nil {
		return "", errors.Wrap(err, "json decoding failed")
	}
	if userResp.Status == "fail" {
		return "", fmt.Errorf("failed to find user, api returned %s", userResp.Status)
	}
	return userResp.User.ID, nil
}

func getPhotosInAlbum(album Album, userID, APIKey string) ([]Photo, error) {
	modifiedUID := strings.Replace(userID, "@", "%40", -1)
	apiURL := fmt.Sprintf("%s?method=%s&api_key=%s&photoset_id=%s&user_id=%s&%s", baseURL, getPhotosFromAlbum, APIKey, album.ID, modifiedUID, commonOptions)

	// Fetch list of photos from API
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, errors.Wrapf(err, "GET request to %s failed", getPhotosFromAlbum)
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
		return nil, errors.Wrap(err, "failed to decode json")
	}
	if photoSetResp.Status != "ok" {
		return nil, fmt.Errorf("api returned bad response: %s", photoSetResp.Status)
	}
	return photoSetResp.Set.Photos, nil
}

func getAllUserPhotos(userID, APIKey string) ([]Photo, error) {
	// Fetch first page
	allPhotos, numPages, err := fetchPhotosByPage(1, userID, APIKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch photos")
	}

	// Fetch all pages
	failedPages := 0
	if numPages < 1 {
		for page := 2; page <= numPages; page++ {
			photosInPage, _, e := fetchPhotosByPage(page, userID, APIKey)
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
	apiURL := fmt.Sprintf("%s?method=%s&api_key=%s&user_id=%s&per_page=500&page=%d&%s", baseURL, getPhotosFromUser, APIkey, modifiedID, page, commonOptions)

	// Fetch JSON from Flickr API
	resp, err := http.Get(apiURL)
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
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to decode json")
	}
	if target.Status != "ok" {
		return nil, 0, fmt.Errorf("unable to get all photos")
	}
	return target.Photos.Photos, target.Photos.Pages, nil
}

func getAlbumsByUserID(userID, APIkey string) ([]Album, error) {
	// Compose request URL
	modifiedID := strings.Replace(userID, "@", "%40", -1)
	apiURL := fmt.Sprintf("%s?method=%s&api_key=%s&user_id=%s&%s", baseURL, getAlbumsFromUserID, APIkey, modifiedID, commonOptions)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, errors.Wrap(err, "GET request failed")
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
		return nil, errors.Wrap(err, "failed to decode json")
	}
	if albumResp.Status != "ok" {
		return nil, fmt.Errorf("non-OK status received from flickr: %s", albumResp.Status)
	}
	return albumResp.Albums.Sets, nil
}
