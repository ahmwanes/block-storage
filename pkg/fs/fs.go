package fs

import (
	"fmt"

	"github.com/ahmwanes/bfs/pkg/block"
)

// InodeCount is the maximum number of files/directories the file system can hold.
const (
	InodeCount = 500
)

// FS represents a simple file system built on top of an encrypted block device.
// It manages the superblock, bitmap, inode table, and data blocks.
type FS struct {
	device *block.BlockDevice // Underlying encrypted block storage
	sb     *SuperBlock        // File system metadata
}

// NewFS creates a new file system on the given block device.
// This formats the device, writing the superblock to block 0.
// WARNING: This will overwrite any existing data on the device.
func NewFS(device *block.BlockDevice) (*FS, error) {
	// Data blocks start after: superblock (1) + bitmap (1) + inode table (10)
	dataStart := 12

	sp := &SuperBlock{
		Magic:           Magic,
		BlockSize:       device.BlockSize(),
		BlockCount:      device.BlockCount(),
		InodeCount:      InodeCount,
		BitmapStart:     1,
		InodeStart:      2,
		DataStart:       uint8(dataStart),
		FreeBlocksCount: device.BlockCount() - uint64(dataStart),
		FreeInodesCount: InodeCount,
	}

	err := device.WriteBlock(0, sp.serialize())
	if err != nil {
		return nil, fmt.Errorf("failed to write superblock: %w", err)
	}
	return &FS{device: device, sb: sp}, nil
}
