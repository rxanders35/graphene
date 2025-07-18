build:
	rm bin/sss
	go build -o bin/sss .

run: build
	./bin/sss --port=:8081 --data-dir=/tmp/volume_data

test:
	go test ./... -v
