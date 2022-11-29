all: wfm

wfm: wfm.go
	go build wfm.go

cross:
	GOOS=linux GOARCH=amd64 go build -a -o wfm-amd64-linux wfm.go
	GOOS=linux GOARCH=arm go build -a -o wfm-arm-linux wfm.go
	GOOS=linux GOARCH=arm64 go build -a -o wfm-arm64-linux wfm.go
	GOOS=darwin GOARCH=amd64 go build -a -o wfm-amd64-macos wfm.go
	GOOS=darwin GOARCH=arm64 go build -a -o wfm-arm64-macos wfm.go
	GOOS=freebsd GOARCH=amd64 go build -a -o wfm-amd64-freebsd wfm.go
	GOOS=openbsd GOARCH=amd64 go build -a -o wfm-amd64-openbsd wfm.go

docker: wfm
	docker build -t tenox7/wfm:latest .

dockerhub:
	docker push tenox7/wfm:latest

gcrio:
	docker tag tenox7/wfm:latest gcr.io/tenox7/wfm
	docker push gcr.io/tenox7/wfm

clean:
	rm -rf wfm-* wfm
