build:
	rm bin/fs
	go build -o bin/fs

run: build
	./bin/fs

test:
	@o test ./... -v
