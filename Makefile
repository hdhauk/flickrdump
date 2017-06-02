crosscompile:
	env GOOS=linux GOARCH=amd64 go build -o bin/flickrdump_linux_amd64
	env GOOS=linux GOARCH=386 go build -o bin/flickrdump_linux_386
	env GOOS=darwin GOARCH=amd64 go build -o bin/flickrdump_darwin_amd64
	env GOOS=windows GOARCH=amd64 go build -o bin/flickrdump_windows_amd64
	env GOOS=windows GOARCH=386 go build -o bin/flickrdump_windows_386
