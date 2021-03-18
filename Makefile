server_src = $(shell find cmd/server -name '*.go')
lib_src = $(shell find hobbyq -name '*.go')

build/hobbyq-server: $(server_src) $(lib_src)
	cd cmd/server && go build -o ../../build/hobbyq-server .
