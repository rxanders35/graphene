package main

// CONSTANTS FOR NEEDLE FORMAT ON DISK

// NEEDLE: MAGICNUMBER|UUID|DATA|DATA_SIZE|CHECKSUM|SIZE

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

/////////////////////////////////////

// CONSTANTS FOR OBJECT INDEX ON DISK

// Object ID field
var idxObjectID [16]byte

// Offset field
var idxOffset uint64

// Size field
var idxSize uint32

// The total size of the Index entry
var idxSizeTotal [28]byte