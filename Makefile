default:
	go install -ldflags "-X main.key=$(apikey)" .
crosscompile:
	env GOOS=linux  go build -o bin/flickrdump_linux -ldflags "-X main.key=$(apikey)"
	env GOOS=darwin  go build -o bin/flickrdump_darwin -ldflags "-X main.key=$(apikey)"
	env GOOS=windows  go build -o bin/flickrdump_windows.exe -ldflags "-X main.key=$(apikey)"
