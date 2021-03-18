// the hobbyq server process

package hobbyq

import (
	"fmt"
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
		fmt.Printf("received client connection: %s\n", conn.RemoteAddr())
		conn.Close()
	}

	return nil
}
