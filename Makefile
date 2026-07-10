VERS    := $(shell git describe --tags --always --abbrev=0)
LDFLAGS := -ldflags "-X main.vers=$(VERS)"

all: wfm

wfm: *.go html/*.html
	go build $(LDFLAGS) .

cross:
	GOOS=linux GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-linux .
	GOOS=linux GOARCH=arm64 go build -a $(LDFLAGS) -o wfm-arm64-linux .
	GOOS=darwin GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-macos .
	GOOS=darwin GOARCH=arm64 go build -a $(LDFLAGS) -o wfm-arm64-macos .
	GOOS=freebsd GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-freebsd .
	GOOS=freebsd GOARCH=arm64 go build -a $(LDFLAGS) -o wfm-arm64-freebsd .
	GOOS=openbsd GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-openbsd .
	GOOS=netbsd GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-netbsd .

docker: wfm
	GOOS=linux GOARCH=amd64 go build -a $(LDFLAGS) -o wfm-amd64-linux .
	GOOS=linux GOARCH=arm64 go build -a $(LDFLAGS) -o wfm-arm64-linux .
	docker buildx build --platform linux/amd64,linux/arm64 -t tenox7/wfm:latest --push .

clean:
	rm -rf wfm-* wfm
