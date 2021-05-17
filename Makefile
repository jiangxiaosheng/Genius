build:
	rm -rf bin; mkdir bin; cd bin; go build -o genius ../cmd/main.go