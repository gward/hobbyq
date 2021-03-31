package main

import (
	"fmt"
	"os"

	"github.com/gward/hobbyq"
)

func main() {
	addr := "localhost:7253"
	err := run(addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(addr string) error {
	client := hobbyq.NewClient(addr)
	err := client.Connect()
	if err != nil {
		return err
	}
	resp, err := client.SendRequest("XMAKE", []string{"foobar"})
	if err != nil {
		return err
	}
	fmt.Printf("resp: status %d, args %q\n", resp.Status, resp.Args)

	resp, err = client.SendRequest("QMAKE", []string{"blah"})
	if err != nil {
		return err
	}
	fmt.Printf("resp: status %d, args %q\n", resp.Status, resp.Args)

	resp, err = client.SendRequest("DUMP", []string{"json"})
	if err != nil {
		return err
	}
	fmt.Printf("resp: status %d, dump %s\n", resp.Status, resp.Args[0])

	return nil
}
