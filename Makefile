server_src = $(shell find cmd/server -name '*.go')
client_src = $(shell find cmd/client -name '*.go')
pb_src = hobbyq/pb/hobbyq.pb.go
lib_src = $(shell find hobbyq -name '*.go') $(pb_src)

protoc_plugin = build/protoc-gen-go
proto = hobbyq/hobbyq.proto

server_bin = build/hobbyq-server
client_bin = build/hobbyq-client

all: $(server_bin) $(client_bin)

$(server_bin): $(server_src) $(lib_src)
	cd cmd/server && go build -o ../../$(server_bin) .

$(client_bin): $(client_src) $(lib_src)
	cd cmd/client && go build -o ../../$(client_bin) .

$(protoc_plugin):
	go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go

$(pb_src): $(proto) $(protoc_plugin)
	protoc --go_out=. $(proto) --plugin=$(protoc_plugin)
