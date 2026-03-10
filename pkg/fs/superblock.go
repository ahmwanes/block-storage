package fs

import "encoding/binary"

// Magic is the file system identifier written at the start of the superblock.
// Used to verify that a block device contains a valid FS_AHMED file system.
var Magic = [8]byte{'F', 'S', '_', 'A', 'H', 'M', 'E', 'D'}

// SuperBlock contains metadata about the entire file system.
// It is stored in block 0 and acts as the "table of contents" for the file system.
// When mounting a file system, the superblock is read first to understand
// the disk layout and available resources.
type SuperBlock struct {
	Magic           [8]byte // File system identifier ("FS_AHMED")
	BlockSize       uint32  // Size of each block in bytes (typically 4096)
	BlockCount      uint64  // Total number of blocks in the file system
	InodeCount      uint64  // Maximum number of inodes (files/directories)
	BitmapStart     uint8   // Block number where the bitmap starts
	InodeStart      uint8   // Block number where the inode table starts
	DataStart       uint8   // Block number where data blocks start
	FreeBlocksCount uint64  // Number of free blocks available
	FreeInodesCount uint64  // Number of free inodes available
}

// serialize converts the SuperBlock struct to bytes for writing to disk.
// The layout is:
//   - Bytes 0-7:   Magic
//   - Bytes 8-11:  BlockSize
//   - Bytes 12-19: BlockCount
//   - Bytes 20-27: InodeCount
//   - Byte 28:     BitmapStart
//   - Byte 29:     InodeStart
//   - Byte 30:     DataStart
//   - Bytes 31-38: FreeBlocksCount
//   - Bytes 39-46: FreeInodesCount
func (sp *SuperBlock) serialize() []byte {
	buf := make([]byte, sp.BlockSize)

	copy(buf[0:8], sp.Magic[:])
	binary.BigEndian.PutUint32(buf[8:12], sp.BlockSize)
	binary.BigEndian.PutUint64(buf[12:20], sp.BlockCount)
	binary.BigEndian.PutUint64(buf[20:28], sp.InodeCount)
	buf[28] = sp.BitmapStart
	buf[29] = sp.InodeStart
	buf[30] = sp.DataStart
	binary.BigEndian.PutUint64(buf[31:39], sp.FreeBlocksCount)
	binary.BigEndian.PutUint64(buf[39:47], sp.FreeInodesCount)

	return buf
}
