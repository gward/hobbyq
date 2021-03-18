package main

import (
	"github.com/gward/hobbyq"
)

func main() {
	server := hobbyq.NewServer("localhost:7253")
	server.Run()
}
