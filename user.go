package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// User is the actual user payload received.
type User struct {
	ID       string `json:"id"`
	NSID     string `json:"nsid"`
	Username Content
}

// UserResp is the full response received from the findByUsername endpoint.
type UserResp struct {
	User   `json:"user"`
	Status string `json:"stat"`
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
