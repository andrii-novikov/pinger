build:
	GOOS=linux GOARCH=arm go build -o bin/raspbery/pinger .
	go build -o bin/mac/pinger .