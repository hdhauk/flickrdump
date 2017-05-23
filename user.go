package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type User struct {
	ID       string `json:"id"`
	NSID     string `json:"nsid"`
	Username Content
}

type UserResp struct {
	User   `json:"user"`
	Status string `json:"stat"`
}

func getUserIDByUsername(username string) (string, error) {
	searchFor := strings.Replace(username, " ", "+", -1)
	req := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.people.findByUsername&api_key=%s&username=%s&format=json&nojsoncallback=1", key, searchFor)

	resp, err := http.Get(req)
	if err != nil {
		mainlogger.Fatalln(err)
	}
	defer resp.Body.Close()
	var u UserResp
	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return "", err
	}
	if u.Status == "fail" {
		return "", fmt.Errorf("failed to find user")
	}
	return u.User.ID, err
}
