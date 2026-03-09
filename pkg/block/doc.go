// Package block provides an encrypted block storage implementation.
//
// It simulates a block device (like a physical disk) using a file as the
// underlying storage. All data blocks are encrypted using AES-256 in CTR mode,
// with keys derived from a password using PBKDF2.
//
// # Disk Layout
//
// The disk image file has the following structure:
//
//	┌─────────────────────────────────────────────────────────┐
//	│ Block 0: Header (unencrypted)                           │
//	│   - Magic bytes: "BFSENC" (6 bytes)                     │
//	│   - Salt for PBKDF2 (32 bytes)                          │
//	│   - Verification hash (32 bytes)                        │
//	├─────────────────────────────────────────────────────────┤
//	│ Block 1: First data block (encrypted)                   │
//	├─────────────────────────────────────────────────────────┤
//	│ Block 2: Second data block (encrypted)                  │
//	├─────────────────────────────────────────────────────────┤
//	│ ...                                                     │
//	└─────────────────────────────────────────────────────────┘
//
// # Encryption
//
// The encryption process uses:
//   - PBKDF2 with SHA-256 (100,000 iterations) for key derivation
//   - AES-256 in CTR mode for block encryption
//   - Unique IV per block (derived from block number)
//
// # Usage
//
// Create a new encrypted block device:
//
//	device, err := block.New("disk.img", block.GB, "mypassword")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer device.Close()
//
// Write data to a block:
//
//	data := []byte("Hello, World!")
//	err = device.WriteBlock(0, data)
//
// Read data from a block:
//
//	data, err := device.ReadBlock(0)
//
// # Block Size
//
// The default block size is 4KB (4096 bytes). Common size constants are
// provided for convenience: KB, MB, GB.
package block
