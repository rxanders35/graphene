package main

import (
	"encoding/binary"
	"io"
)

// ///////////////////////////////////
const (
	registerMsg = 1
	storeMsg    = 2
	successMsg  = 3
	msgSize     = 4
	pingMsg     = 4
	pongMsg     = 5
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

func encodeMessage(msgType byte, payload []byte) []byte {
	msgLength := uint32(msgTypeSize + len(payload))
	buf := make([]byte, msgSize+msgLength)
	binary.BigEndian.PutUint32(buf[:msgSize], msgLength)
	buf[4] = msgType
	copy(buf[5:], payload)
	return buf
}

func decodeMessage(r io.Reader) (msgType byte, payload []byte, err error) {
	lenBuf := make([]byte, 4)
	if _, err = io.ReadFull(r, lenBuf); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	buf := make([]byte, length)
	if _, err = io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	t := buf[0]
	p := buf[1:]
	return t, p, nil
}

func encodeRegister(addr string) []byte {
	payload := []byte(addr)
	p := encodeMessage(registerMsg, payload)
	return p
}

func encodeStore(uuid [16]byte, data []byte) []byte {
	payload := make([]byte, 16+len(data))
	copy(payload[:16], uuid[:])
	copy(payload[16:], data)
	p := encodeMessage(storeMsg, payload)
	return p
}

func encodeSuccess(resp string) []byte {
	payload := []byte(resp)
	p := encodeMessage(successMsg, payload)
	return p
}

func encodePing() []byte {
	p := encodeMessage(pingMsg, []byte{})
	return p
}
