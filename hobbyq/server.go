// the hobbyq server process

package hobbyq

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	address string				// host:port
	exchanges map[string] *Exchange
	queues map[string] *Queue
}

// Return a new Server struct. It's ready to listen for connections but
// not actually listening. Call Run() to make that happen.
func NewServer(addr string) *Server {
	return &Server{
		addr,
		make(map[string] *Exchange),
		make(map[string] *Queue),
	}
}

// Start listening for connections. Run forever waiting for clients
// and handling them. Never returns, unless there is an error.
func (server *Server) Run() error {
	listener, err := net.Listen("tcp", server.address)
	defer listener.Close()
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

	err := handshake(conn)
	if err != nil {
		log.Printf("error: %s", err)
		return
	}
	for {
		err = server.processCommand(conn)
		if err == io.EOF {
			log.Printf("client disconnected")
			break
		} else if err != nil {
			log.Printf("error: %s", err)
			conn.Write(RESP_BAD_SYNTAX)
			break
		}
	}
}

func handshake(conn io.ReadWriter) error {
	// The first handshake line from the client is fixed length:
	//   HQ 0000\n
	// where 0000 is the client's protocol version (hex-encoded,
	// unsigned 16-bit integer)
	resp := RESP_BAD_SYNTAX			// assume the worst
	respond := func() { conn.Write(resp) }
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
	resp = RESP_OK
	return nil
}

// Read and process a single command. If client input is malformed and
// cannot be parsed, return an error. Otherwise respond to the client,
// possibly with an error response, and return nil.
func (server *Server) processCommand(conn io.ReadWriter) error {
	cmd, err := readString(conn)
	log.Printf("cmd = %q, err = %v", cmd, err)
	if err != nil {
		return err
	}
	args, err := readStringArray(conn)
	log.Printf("args = %q, err = %v", args, err)
	if err != nil {
		return err
	}
	cfunc := COMMAND_FUNC[cmd]
	if cfunc == nil {
		conn.Write(RESP_UNKNOWN_CMD)
		return nil
	}
	resp, err := cfunc(server, args)
	if len(resp) > 0 {
		conn.Write(resp)
	}
	return err
}

func readString(conn io.Reader) (value string, err error) {
	length, err := readLengthByte(conn)
	if err != nil {
		return
	}
	data := make([]byte, length)
	n, err := conn.Read(data)
	log.Printf("string: n=%v, err=%v, data=%q", n, err, data)
	if err != nil {
		return
	}
	if n < length {
		return value, io.ErrUnexpectedEOF
	}
	value = string(data)

	return
}

func readStringArray(conn io.Reader) (values []string, err error) {
	length, err := readLengthByte(conn)
	if err != nil {
		return
	}

	values = make([]string, length)
	for i := 0; i < length; i++ {
		var val string
		val, err = readString(conn)
		if err != nil {
			return
		}
		values[i] = val
	}
	return
}

func readLengthByte(conn io.Reader) (value int, err error) {
	var lengthBuf [1]byte
	n, err := conn.Read(lengthBuf[0:1])
	log.Printf("length byte: n=%v, err=%v, buf=%q", n, err, lengthBuf)
	if err != nil {
		return
	}
	if n < 1 {
		err = io.ErrUnexpectedEOF
		return
	}
	value = int(lengthBuf[0])
	return
}

// Conditionally create an exchange: if it already exists, do nothing
// and return RESP_OK. If it doesn't exist, create it and return
// RESP_CREATED.
func cmd_xmake(server *Server, args []string) (resp []byte, err error) {
	resp = RESP_BAD_ARGS
	if len(args) != 1 {
		err = errors.New("XMAKE: require exactly one argument")
		return
	}
	name := args[0]
	exchange := server.exchanges[name]
	if exchange == nil {
		server.exchanges[name] = NewExchange(name)
		resp = RESP_CREATED
	} else {
		resp = RESP_OK
	}
	return
}

// Conditionally create a queue.
func cmd_qmake(server *Server, args []string) (resp []byte, err error) {
	// Hmmmm: this is basically identical to cmd_xmake().
	resp = RESP_BAD_ARGS
	if len(args) != 1 {
		err = errors.New("QMAKE: require exactly one argument")
		return
	}
	name := args[0]
	queue := server.queues[name]
	if queue == nil {
		server.queues[name] = NewQueue(name)
		resp = RESP_CREATED
	} else {
		resp = RESP_OK
	}
	return
}

var COMMAND_FUNC = map[string] func(*Server, []string) ([]byte, error){
	"XMAKE": cmd_xmake,
	"QMAKE": cmd_qmake,
}
