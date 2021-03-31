package hobbyq

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"google.golang.org/protobuf/proto"
)

// Protocol definitions needed by both server and clients.
// (Hmmmm: these need to be available in a client library, so
// probably a separate package!)

import (
	"github.com/gward/hobbyq/pb"
)

const (
	STATUS_OK = 200
	STATUS_CREATED = 201
	STATUS_BAD_SYNTAX = 400
	STATUS_UNKNOWN_CMD = 405
	STATUS_BAD_ARGS = 406
	STATUS_UNSUPPORTED_VERSION = 451
	STATUS_INTERNAL_ERR = 500
)

func NewResponse(status uint32, args ...string) *pb.Response {
	return &pb.Response{Status: status, Args: args}
}

var (
	RESP_OK = NewResponse(STATUS_OK)
	RESP_CREATED = NewResponse(STATUS_CREATED)
	RESP_BAD_SYNTAX = NewResponse(STATUS_BAD_SYNTAX)
	RESP_UNKNOWN_CMD = NewResponse(STATUS_UNKNOWN_CMD)
	RESP_BAD_ARGS = NewResponse(STATUS_BAD_ARGS)
	RESP_UNSUPPORTED_VERSION = NewResponse(STATUS_UNSUPPORTED_VERSION)
	RESP_INTERNAL_ERR = NewResponse(STATUS_INTERNAL_ERR)
)


func sendRequest(conn io.ReadWriter, request *pb.Request) error {
	return sendMessage(conn, request)
}

func sendResponse(conn io.Writer, resp *pb.Response) error {
	return sendMessage(conn, resp)
}

func readRequest(conn io.Reader) (pbreq *pb.Request, err error) {
	buf, err := readMessage(conn)
	log.Printf("buf = %q, err = %v", buf, err)
	if err != nil {
		return
	}

	pbreq = &pb.Request{}
	err = proto.Unmarshal(buf, pbreq)
	log.Printf("readRequest: pbreq = %v, err = %v", pbreq, err)
	return
}

func readResponse(conn io.Reader) (resp *Response, err error) {
	buf, err := readMessage(conn)
	if err != nil {
		return
	}

	pbresp := &pb.Response{}
	err = proto.Unmarshal(buf, pbresp)
	if err != nil {
		return
	}
	log.Printf("read and decoded server response: %v", pbresp)
	resp = &Response{pbresp.Status, pbresp.Args}
	return
}

func sendMessage(conn io.Writer, msg proto.Message) (err error) {
	maxBytes := 64 * 1024 - 1
	buf, err := proto.Marshal(msg)
	if err == nil && len(buf) > maxBytes {
		err = fmt.Errorf("Request message too big (max %d bytes)", maxBytes)
	}
	if err != nil {
		return
	}

	sizeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBuf, uint16(len(buf)))

	nbytes, err := conn.Write(sizeBuf)
	if err == nil && nbytes < 2 {
		err = errors.New("short write")
	}
	if err != nil {
		return
	}

	nbytes, err = conn.Write(buf)
	if err == nil && nbytes < len(buf) {
		err = errors.New("short write")
	}
	return
}

func readMessage(conn io.Reader) (buf []byte, err error) {
	length, err := readLength(conn)
	if err != nil {
		return
	}

	buf = make([]byte, length)
	nbytes, err := conn.Read(buf)
	log.Printf("string: nbytes=%d, err=%v, data=%q", nbytes, err, buf)
	if err == nil && nbytes < int(length) {
		err = io.ErrUnexpectedEOF
	}
	return
}

func readLength(conn io.Reader) (value uint16, err error) {
	buf := make([]byte, 2)
	nbytes, err := conn.Read(buf)
	log.Printf("readLength: nbytes=%d, err=%v, buf=%q", nbytes, err, buf)
	if err == nil && nbytes < 2 {
		err = io.ErrUnexpectedEOF
	}
	if err == nil {
		value = binary.BigEndian.Uint16(buf)
	}
	return
}
