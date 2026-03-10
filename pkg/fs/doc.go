// Package fs provides a simple file system implementation on top of the block storage layer.
//
// This file system uses the block.BlockDevice as its underlying storage,
// organizing data into a structured format with a superblock, bitmap, inode table,
// and data blocks.
//
// # Disk Layout
//
// The file system organizes the block device as follows:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│ Block 0: Superblock                                             │
//	│   - Magic bytes: "FS_AHMED"                                     │
//	│   - Block size, block count, inode count                        │
//	│   - Pointers to bitmap, inode table, data blocks                │
//	├─────────────────────────────────────────────────────────────────┤
//	│ Block 1: Block Bitmap                                           │
//	│   - Tracks which blocks are free (0) or used (1)                │
//	├─────────────────────────────────────────────────────────────────┤
//	│ Blocks 2-11: Inode Table                                        │
//	│   - Fixed-size entries for file/directory metadata              │
//	│   - Each inode contains: type, size, block pointers, timestamps │
//	├─────────────────────────────────────────────────────────────────┤
//	│ Blocks 12+: Data Blocks                                         │
//	│   - Actual file and directory contents                          │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Components
//
// Superblock: The file system's "table of contents". Contains metadata about
// the entire file system including where each section starts and how much
// space is available.
//
// Bitmap: A bit array where each bit represents one block. 0 means the block
// is free, 1 means it's in use. Used to quickly find free blocks when creating
// or extending files.
//
// Inode: Short for "index node". Stores metadata about a file or directory:
//   - Type (file or directory)
//   - Size in bytes
//   - Block pointers (which blocks contain the data)
//   - Timestamps (created, modified)
//
// Directory: A special type of file that maps filenames to inode numbers.
//
// # Usage
//
// Create a new file system on a block device:
//
//	device, _ := block.New("disk.img", block.GB, "password")
//	fs, err := fs.NewFS(device)
//
// # Limitations
//
//   - Maximum file size: 48KB (12 direct blocks × 4KB)
//   - Maximum files: 500 (configurable via InodeCount)
//   - Single-level directory structure (initially)
package fs
