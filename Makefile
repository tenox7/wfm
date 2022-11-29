all: wfm

wfm: *.go
	go build .

cross:
	GOOS=linux GOARCH=amd64 go build -a -o wfm-amd64-linux .
	GOOS=linux GOARCH=arm go build -a -o wfm-arm-linux .
	GOOS=linux GOARCH=arm64 go build -a -o wfm-arm64-linux .
	GOOS=darwin GOARCH=amd64 go build -a -o wfm-amd64-macos .
	GOOS=darwin GOARCH=arm64 go build -a -o wfm-arm64-macos .
	GOOS=freebsd GOARCH=amd64 go build -a -o wfm-amd64-freebsd .
	GOOS=openbsd GOARCH=amd64 go build -a -o wfm-amd64-openbsd .

docker: wfm
	strip wfm
	cp wfm service/docker/
	docker build -t tenox7/wfm:latest service/docker/

dockerhub:
	docker push tenox7/wfm:latest

gcrio:
	docker tag tenox7/wfm:latest gcr.io/tenox7/wfm
	docker push gcr.io/tenox7/wfm

clean:
	rm -rf wfm-* wfm
