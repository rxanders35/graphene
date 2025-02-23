package main

// CONSTANTS FOR NEEDLE FORMAT ON DISK

// NEEDLE: MAGICNUMBER|UUID|DATA|CHECKSUM|SIZE

// Size of Needle's magic number
var needleMagicSize uint16

// Size of Needle's UUID
var needleIDSize [16]byte

// Size of Needle's internal data
var needleDataSize uint32

// Size of Needle's checksum
var needleChecksum uint32

// The total size of the Needle
var needleSizeTotal [26]byte

// The Needle magic number literal
var needleMagicVal = 0xCAFE