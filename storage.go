package main

// CONSTANTS FOR NEEDLE FORMAT ON DISK

// NEEDLE: MAGICNUMBER|UUID|DATA|DATA_SIZE|CHECKSUM|SIZE

// Size of Needle's magic number
const needleMagicSize = 2

// Size of Needle's UUID
const needleIDSize = 16

// Size of Needle's internal data
const needleDataSize = 4

// Size of Needle's checksum
const needleChecksum = 4

// The total size of the Needle
const needleSizeTotal = 26

// The Needle magic number literal
const needleMagicVal = 0xCAFE

/////////////////////////////////////

// CONSTANTS FOR OBJECT INDEX ON DISK

// Object ID field
const idxObjectID = 16

// Offset field
const idxOffset = 8

// Size field
const idxSize = 4

// The total size of the Index entry
const idxSizeTotal = 28