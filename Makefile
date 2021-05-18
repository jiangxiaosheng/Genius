build:
	rm -rf bin; mkdir bin; cd bin; go build -o genius ../cmd/main.go

image: build
	docker rmi genius:latest; docker build -t genius:latest .