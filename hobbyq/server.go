// the hobbyq server process

package hobbyq

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/gward/hobbyq/pb"
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

func (server *Server) MarshalJSON() ([]byte, error) {
	exchanges := make([]*Exchange, 0, len(server.exchanges))
	for _, val := range server.exchanges {
		exchanges = append(exchanges, val)
	}
	queues := make([]*Queue, 0, len(server.queues))
	for _, val := range server.queues {
		queues = append(queues, val)
	}

	var output = map[string] interface{}{
		"exchanges": exchanges,
		"queues": queues,
	}
	return json.Marshal(output)
}

// Start listening for connections. Run forever waiting for clients
// and handling them. Never returns, unless there is an error.
func (server *Server) Run() error {
	listener, err := net.Listen("tcp", server.address)
	defer listener.Close()
	if err != nil {
		return err
	}
	log.Printf("Listening on %v", listener.Addr())
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

	resp, err := handshake(conn)
	if err != nil {
		log.Printf("error: %s", err)
		return
	}
	err = sendResponse(conn, resp)
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
			break
		}
	}
}

func handshake(conn io.ReadWriter) (resp *pb.Response, err error) {
	// The first handshake line from the client is fixed length:
	//   HQ 0000\n
	// where 0000 is the client's protocol version (hex-encoded,
	// unsigned 16-bit integer)
	resp = RESP_BAD_SYNTAX			// assume the worst

	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	if n < 8 {
		err = fmt.Errorf("incomplete client handshake: %q", buf)
		return
	}
	if !(bytes.Equal(buf[0:3], []byte("HQ ")) && buf[7] == '\n') {
		err = fmt.Errorf("invalid client handshake: %q", buf)
		return
	}
	clientVersion := make([]byte, 2)
	n, err = hex.Decode(clientVersion, buf[3:7])
	if n != 2 || err != nil {
		err = fmt.Errorf("invalid client version: %q (%s)", buf[3:7], err)
		return
	}
	if !bytes.Equal(clientVersion, []byte("\x00\x01")) {
		resp = RESP_UNSUPPORTED_VERSION
		err = fmt.Errorf("unsupported client version: %q", clientVersion)
		return
	}

	// Successful handshake.
	resp = RESP_OK
	err = nil
	return
}

// Read and process a single command. If client input is malformed and
// cannot be parsed, return an error. Otherwise respond to the client,
// possibly with an error response, and return nil.
func (server *Server) processCommand(conn io.ReadWriter) error {

	pbreq, err := readRequest(conn)
	if err != nil {
		return err
	}

	cfunc := COMMAND_FUNC[pbreq.Command]
	if cfunc == nil {
		err = sendResponse(conn, RESP_UNKNOWN_CMD)
		return err
	}
	resp, err := cfunc(server, pbreq.Args)
	if err != nil {
		log.Printf("command %s: error: %s (sending response: %v)",
			pbreq.Command, err, resp)
	}
	if resp != nil {
		log.Printf("sending resp %v", resp)
		err = sendResponse(conn, resp)
	}
	return err
}

// Conditionally create an exchange: if it already exists, do nothing
// and return RESP_OK. If it doesn't exist, create it and return
// RESP_CREATED.
func cmd_xmake(server *Server, args []string) (resp *pb.Response, err error) {
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
func cmd_qmake(server *Server, args []string) (resp *pb.Response, err error) {
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

// Dump the state of server to a string. Format of the string is
// specified as the first arugment: "json" or "text".
func cmd_dump(server *Server, args []string) (resp *pb.Response, err error) {
	resp = RESP_BAD_ARGS
	if len(args) != 1 {
		err = errors.New("DUMP: require exactly one argument")
		return
	}
	format := args[0]
	if format != "json" {
		err = errors.New("DUMP: unsupported format (allowed: json)")
		return
	}
	buf, err := json.Marshal(server)
	if err != nil {
		resp = RESP_INTERNAL_ERR
		return
	}
	if len(buf) > 65535 {
		resp = RESP_INTERNAL_ERR
		err = errors.New("DUMP: server state too long")
		return
	}
	log.Printf("DUMP buf = %q", buf)
	resp = NewResponse(200, string(buf))
	log.Printf("DUMP resp = %q", resp)

	return
}

var COMMAND_FUNC = map[string] func(*Server, []string) (*pb.Response, error){
	"XMAKE": cmd_xmake,
	"QMAKE": cmd_qmake,
	"DUMP": cmd_dump,
}
