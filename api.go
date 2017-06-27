package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Photo contain both id and title of a Flickr photo
type Photo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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
