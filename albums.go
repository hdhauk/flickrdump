package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AlbumsResp is the full resonse received from the photosets.getList endpoint.
type AlbumsResp struct {
	Albums Albums `json:"photosets"`
	Status string `json:"stat"`
}

// Albums is the object holding the album array with its length.
type Albums struct {
	Length int     `json:"total"`
	Sets   []Album `json:"photoset"`
}

// Album is the object containing information on one album.
type Album struct {
	ID          string `json:"id"`
	Title       Content
	Description Content
}

func getAlbumsByUser(userID, APIkey string) ([]Album, error) {
	// Compose request URL
	modifiedID := strings.Replace(userID, "@", "%40", -1)
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getList&api_key=%s&user_id=%s&format=json&nojsoncallback=1", APIkey, modifiedID)

	resp, err := http.Get(req)
	if err != nil {
		mainlogger.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response
	var a AlbumsResp
	err = json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		return nil, err
	}
	if a.Status != "ok" {
		return nil, fmt.Errorf("unable to get albums")
	}
	return a.Albums.Sets, nil
}
