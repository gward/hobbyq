package hobbyq

// Protocol definitions needed by both server and clients.
// (Hmmmm: these need to be available in a client library, so
// probably a separate package!)

const RESP_HANDSHAKE_OK = "200\x00"
const RESP_BAD_HANDSHAKE = "450\x00"
const RESP_UNSUPPORTED_VERSION = "451\x00"
