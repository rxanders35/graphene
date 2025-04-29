package main

const (
	registerMsg = 1
	storeMsg    = 2
	successMsg  = 3
	msgSize     = 4
	msgTypeSize = 1
)

// TCP FORMAT : LENGTH(32BYTES) | MSGTYPE(1BYTE) | PAYLOAD
// MESSAGE TYPES
// (1) REGISTER
// - Payload: Volume Server HTTP address
// (2) STORE
// - Payload: 16 byte UUID + Object Data
// (3) SUCCESS
// - Payload: Response String
