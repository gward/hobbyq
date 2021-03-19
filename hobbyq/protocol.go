package hobbyq

// Protocol definitions needed by both server and clients.
// (Hmmmm: these need to be available in a client library, so
// probably a separate package!)

var (
	RESP_OK = []byte("200\x00")
	RESP_CREATED = []byte("201\x00")
	RESP_BAD_SYNTAX = []byte("400\x00")
	RESP_UNKNOWN_CMD = []byte("405\x00")
	RESP_BAD_ARGS = []byte("406\x00")
	RESP_UNSUPPORTED_VERSION = []byte("451\x00")
)
