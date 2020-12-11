default: build server

build:
	go build

.PHONY: server
server:
	./simpledb

.PHONY: client
client: build
	./simpledb -client

.PHONY: clean
clean:
	rm simpledb
	find . -name *.out -or -name *.log -or -name .*.swp -or -name .*.swo -or -name .DS_Store -or -name .swp | xargs -n 1 rm

