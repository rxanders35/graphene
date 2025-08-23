package needle

/////////////////////////////////////

// CONSTANTS FOR NEEDLE FORMAT ON DISK

/////////////////////////////////////

// NEEDLE: MAGICNUMBER|UUID|SIZE|DATA|CHECKSUM

const (
	// Size of Needle's magic number
	NeedleMagicSize = 2

	// Size of Needle's UUID
	NeedleIDSize = 16

	// Size of Needle's blob data payload
	NeedleDataSize = 4

	// Size of Needle's checksum
	NeedleChecksum = 4

	// The total fixed overhead of the Needle
	NeedleFixedPortion = 26

	// The Needle magic number literal
	NeedleMagicVal uint16 = 0xCAFE
)

/////////////////////////////////////

// CONSTANTS FOR OBJECT INDEX ON DISK

/////////////////////////////////////

// IDX: OBJECT_ID|OFFSET|SIZE
const (
	// Object ID field
	IdxObjectID = 16

	// Offset field
	IdxOffset = 8

	// Size field
	IdxSize = 4

	// The total size of the Index entry
	IdxEntryTotalSize = 28
)

type IndexEntry struct {
	ID     [16]byte
	Offset uint64
	Size   uint32
}

/////////////////////////////////////

// CONSTANTS FOR NAMING STANDARDS ON DISK

// ///////////////////////////////////
const (
	// Volume server classification
	VolumeFilePrefix = "volume_"

	// Data file suffix
	DataFileExtension = ".dat"

	// Index file suffix
	IdxFileExtension = ".idx"

	// Directory
	DataDir = "data/"
)
