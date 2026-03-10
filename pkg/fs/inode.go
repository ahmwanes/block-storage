// Package fs implements a simple file system on top of a block device.
package fs

import "encoding/binary"

// Inode types
const (
	InodeTypeFile = 1 // Regular file
	InodeTypeDir  = 2 // Directory
)

// Inode represents metadata about a file or directory.
// It stores everything about a file EXCEPT its name (which is in the directory).
//
// The Blocks array contains pointers to data blocks where the file content is stored.
// With 12 direct block pointers and 4KB blocks, max file size is 48KB.
// For larger files, indirect blocks would be needed (not implemented).
type Inode struct {
	Type       uint8      // File type: InodeTypeFile or InodeTypeDir. 1=file, 2=directory
	Size       uint64     // File size in bytes
	BlockCount uint8      // Number of blocks used (0-12)
	Blocks     [12]uint64 // Direct block pointers
	CreatedAt  int64      // Unix timestamp of creation
	UpdatedAt  int64      // Unix timestamp of last modification
}

// Inode size calculation:
// Type(1) + Size(8) + BlockCount(1) + Blocks(12*8=96) + CreatedAt(8) + UpdatedAt(8) = 122 bytes
// Padded to 128 bytes for alignment
const InodeSize = 128

// serialize converts the Inode struct to bytes for writing to disk.
// The layout is:
//   - Byte 0:      Type
//   - Bytes 1-8:   Size
//   - Byte 9:      BlockCount
//   - Bytes 10-17: Blocks[0]
//   - Bytes 18-25: Blocks[1]
//   - ...
//   - Bytes 110-117: CreatedAt
//   - Bytes 118-125: UpdatedAt
func (i *Inode) serialize() []byte {
	buf := make([]byte, InodeSize)

	binary.BigEndian.PutUint16(buf, uint16(i.Type))
	binary.BigEndian.PutUint64(buf[1:9], i.Size)
	buf[9] = i.BlockCount
	// TODO: add block pointers
	binary.BigEndian.PutUint64(buf[110:118], uint64(i.CreatedAt))
	binary.BigEndian.PutUint64(buf[118:126], uint64(i.UpdatedAt))

	return buf
}
