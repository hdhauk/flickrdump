# flickrdump

### Highlights
* Download all photos from a specified account, also those outside of albums.
* Arbitrarily many parallel downloads
* Sorts into folder per album
* Binaries with API key included

![](screen.gif)

### Description
Basically a script put together in a hurry to dump all the pictures belonging to single Flickr user in the original size. I have done no testing what so ever, and pretty much only made this to dump all the NASA Apollo pictures. All photos in an album is downloaded concurrently so beware it will likely hog your bandwidth if you are using more than 4-5. Default number of concurrent http calls is 4. It also skips already downloaded files so you can resume easily at a later time.

### Installation
Either trust me and use my api-key with the binaries provided:
* [Windows](./bin/flickrdump_windows.exe)
* [MacOS](./bin/flickrdump_darwin)
* [Linux](./bin/flickrdump_linux)

OR get your own key and build from the source:
```
go get -u github.com/hdhauk/flickrdump
make apikey=<your-api-key-goes-here>
```
You can get your Flickr API key [here](https://www.flickr.com/services/api/misc.api_keys.html).

### Usage
```
$ flickrdump -h
Usage of flickrdump:
  -n int
        Number of concurrent downloads. (default 4)
  -noalbums
        Ignore photos that are in one or more albums.
  -nodownload
        Only scan photos, skip actual download. Useful to check for number of photos
  -nounsorted
        Ignore photos not found in any albums.
  -out string
        Destination folder. (default "/Users/halvor/Documents/Go/src/github.com/hdhauk/flickrdump")
  -url string
        URL to the profile page of the account to dump from.
  -user string
        Username for account to dump from. Note that this is not always what appares on the profile page.
```

### Examples
```
flickrdump -nounsorted -user="Apollo Image Gallery"
flickrdump -url=https://www.flickr.com/photos/spacex -noalbums -out=$HOME/Pictures/
```
