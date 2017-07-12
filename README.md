# flickrdump

### Highlights
* Download all photos from a specified account, also those outside of albums.
* Arbitrarily many parallel downloads
* Sorts into folder per album
* Binaries with API key included
* No external dependencies

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
NAME:
   flickrdump - Download photos from Flickr, the fast way!

USAGE:
   flickrdump [global options] command [command options] [arguments...]

COMMANDS:
     username  The username to download from
     url       The url of the user to download from
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --ignore-albums    Do not download photos that are sorted into albums
   --ignore-unsorted  Do not download photos that aren't part of any albums
   --threads value    Number of concurrent downloads (default: 4)
   --no-download      Only scan user, no files will be downloaded.
   --help, -h         show help
```

### Examples
```
flickrdump -ignore-albums url https://www.flickr.com/photos/spacex
flickrdump username "Apollo Image Gallery"
```

