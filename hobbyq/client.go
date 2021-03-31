package hobbyq

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/gward/hobbyq/pb"
)

type Client struct {
	server string				// host:port
	conn net.Conn
}

type Response struct {
	Status uint32
	Args []string
}

func NewClient(server string) *Client {
	return &Client{server, nil}
}

func (client *Client) Connect() error {
	conn, err := net.Dial("tcp", client.server)
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", conn.RemoteAddr())

	err = client.handshake(conn)
	if err != nil {
		conn.Close()
		return err
	}

	client.conn = conn
	return nil
}

func (client *Client) handshake(conn io.ReadWriter) error {
	// This client implements protocol version 1.
	greeting := []byte("HQ 0001\n")
	nbytes, err := conn.Write(greeting)
	if err != nil {
		return err
	}
	if nbytes < len(greeting) {
		return fmt.Errorf("handshake error: short write (only %d/%d bytes)",
			nbytes, len(greeting))
	}
	// log.Printf("successfully sent %d byte greeting: %q", nbytes, greeting)

	// Expect server to reply with RESP_OK.
	resp, err := readResponse(conn)
	if err != nil {
		return err
	}
	if resp.Status == 451 {
		return errors.New("server does not support this protocol version")
	} else if resp.Status != 200 {
		return fmt.Errorf("server handshake failed: %v", resp)
	}
	return nil
}

func (client *Client) SendRequest(command string, args []string) (resp *Response, err error) {
	request := &pb.Request{}
	request.Command = command
	request.Args = args

	err = sendRequest(client.conn, request)
	if err != nil {
		return
	}
	resp, err = readResponse(client.conn)
	return
}
