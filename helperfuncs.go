package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

func notAlreadyDownloaded(done, all []Photo) []Photo {
	photoMap := make(map[string]Photo)

	// Add all photos to map
	for _, p := range all {
		photoMap[p.ID] = p
	}

	// Remove all downloaded photos.
	for _, p := range done {
		delete(photoMap, p.ID)
	}

	// Convert back to slice.
	var notDownloaded []Photo
	for _, p := range photoMap {
		notDownloaded = append(notDownloaded, p)
	}

	// Sort photos to avoid duplicate saves if partially downloaded before
	sort.Sort(PhotoSet(notDownloaded))

	return notDownloaded
}

func createFolder(rootPath, name string) string {
	cleanName := sanitize(name)
	folderPath := path.Join(rootPath, cleanName)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.Mkdir(folderPath, 0700)
	}
	return folderPath
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
		{"\"", "-"},
	}

	clean := dirty
	for _, v := range illegalSubStrings {
		clean = strings.Replace(clean, v.Illegal, v.ReplaceWith, -1)
	}
	return clean
}

func getUserArgs() (user string, workers int) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Flickr user to download all albums from: ")
	user, _ = reader.ReadString('\n')
	user = strings.TrimSuffix(user, "\n")

	fmt.Print("Number of parallell downloads: ")
	fmt.Scanf("%d", &workers)
	fmt.Println()

	return
}
