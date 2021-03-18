// the hobbyq server process

package hobbyq

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	address string				// host:port
}

// Return a new Server struct. It's ready to listen for connections but
// not actually listening. Call Run() to make that happen.
func NewServer(addr string) *Server {
	return &Server{addr}
}

// Start listening for connections. Run forever waiting for clients
// and handling them. Never returns, unless there is an error.
func (server *Server) Run() error {
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return nil
		}
		log.Printf("received client connection: %s", conn.RemoteAddr())
		go server.handleConnection(conn)
	}

	return nil
}

// Handle a single client connection, for however long that might take.
func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	err := server.handshake(conn)
	if err != nil {
		log.Printf("error: %s", err)
		return
	}
}

func (server *Server) handshake(conn io.ReadWriter) error {
	// The first handshake line from the client is fixed length:
	//   HQ 0000\n
	// where 0000 is the client's protocol version (hex-encoded,
	// unsigned 16-bit integer)
	resp := RESP_BAD_HANDSHAKE			// assume the worst
	respond := func() { conn.Write([]byte(resp)) }
	defer respond()

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	if n < 8 {
		return fmt.Errorf("incomplete client handshake: %q", buf)
	}
	if !(bytes.Equal(buf[0:3], []byte("HQ ")) && buf[7] == '\n') {
		return fmt.Errorf("invalid client handshake: %q", buf)
	}
	clientVersion := make([]byte, 2)
	n, err = hex.Decode(clientVersion, buf[3:7])
	if n != 2 || err != nil {
		return fmt.Errorf("invalid client version: %q (%s)", buf[3:7], err)
	}
	if !bytes.Equal(clientVersion, []byte("\x00\x00")) {
		resp = RESP_UNSUPPORTED_VERSION
		return fmt.Errorf("unsupported client version: %q", clientVersion)
	}

	// Successful handshake.
	resp = RESP_HANDSHAKE_OK
	return nil
}
