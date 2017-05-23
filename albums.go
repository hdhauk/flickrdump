package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AlbumsResp struct {
	Albums Albums `json:"photosets"`
	Status string `json:"stat"`
}

type Albums struct {
	Lenght int     `json:"total"`
	Sets   []Album `json:"photoset"`
}

type Album struct {
	ID          string `json:"id"`
	Title       Content
	Description Content
}

func getAlbumsByUser(userID string) ([]Album, error) {
	userID = strings.Replace(userID, "@", "%40", -1)
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photosets.getList&api_key=%s&user_id=%s&format=json&nojsoncallback=1", key, userID)

	resp, err := http.Get(req)
	if err != nil {
		mainlogger.Fatalln(err)
	}
	defer resp.Body.Close()
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
