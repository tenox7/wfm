all: wfm

wfm: *.go
	go build .

cross:
	GOOS=linux GOARCH=amd64 go build -a -o wfm-amd64-linux .
	GOOS=linux GOARCH=arm64 go build -a -o wfm-arm64-linux .
	GOOS=darwin GOARCH=amd64 go build -a -o wfm-amd64-macos .
	GOOS=darwin GOARCH=arm64 go build -a -o wfm-arm64-macos .
	GOOS=freebsd GOARCH=amd64 go build -a -o wfm-amd64-freebsd .
	GOOS=openbsd GOARCH=amd64 go build -a -o wfm-amd64-openbsd .

docker: wfm
	GOOS=linux GOARCH=amd64 go build -a -o wfm-amd64-linux .
	GOOS=linux GOARCH=arm64 go build -a -o wfm-arm64-linux .
	docker buildx build --platform linux/amd64,linux/arm64 -t tenox7/wfm:latest --push .

clean:
	rm -rf wfm-* wfm
