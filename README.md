# flickrdump

Basically a script put together in a hurry to dump all the pictures beloning to single Flickr user in the original size. I have done no testing what so ever, and pretty much only made this to dump all the NASA Apollo pictures. All photos in an album is downloaded concurrently so beware it will likely hog your bandwidth.

### Installation
```
go get -u github.com/hdhauk/flickrdump
```
You will also need a Flickr API key [here](https://www.flickr.com/services/api/misc.api_keys.html).

### Example
```
flickrdump -key <your-key> -u "Apollo Image Gallery"
```
