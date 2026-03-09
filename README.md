# Block Storage (BFS)

An encrypted block storage implementation in Go. This project simulates a block device (like a physical disk) using a file as the underlying storage, with AES-256 encryption for all data blocks.

## Overview

This project was built as a learning exercise to understand:
- How block storage works at a low level
- The difference between block storage and file storage
- How disk encryption works (similar to FileVault, BitLocker, LUKS)
- Key derivation using PBKDF2
- AES encryption in CTR mode

## Features

- **Virtual Block Device**: Creates a file that acts as a raw disk with fixed-size blocks (4KB)
- **AES-256 Encryption**: All data blocks are encrypted using AES-256-CTR
- **Password-Based Key Derivation**: Uses PBKDF2 with SHA-256 (100,000 iterations) to derive encryption keys
- **Password Verification**: Validates password on open without exposing the key

## Disk Layout

```
┌─────────────────────────────────────────────────────────┐
│ Block 0: Header (unencrypted)                           │
│   - Magic bytes: "BFSENC" (6 bytes)                     │
│   - Salt for PBKDF2 (32 bytes)                          │
│   - Verification hash (32 bytes)                        │
├─────────────────────────────────────────────────────────┤
│ Block 1: First data block (encrypted)                   │
├─────────────────────────────────────────────────────────┤
│ Block 2: Second data block (encrypted)                  │
├─────────────────────────────────────────────────────────┤
│ ...                                                     │
└─────────────────────────────────────────────────────────┘
```

## Installation

```bash
go get github.com/ahmwanes/bfs
```

## Usage

### Create a New Encrypted Disk

```go
package main

import (
    "fmt"
    "github.com/ahmwanes/bfs/pkg/block"
)

func main() {
    // Create a 1GB encrypted disk
    device, err := block.New("disk.img", block.GB, "mypassword")
    if err != nil {
        panic(err)
    }
    defer device.Close()

    // Write data to block 0
    data := []byte("Hello, Block Storage!")
    err = device.WriteBlock(0, data)
    if err != nil {
        panic(err)
    }

    // Read data back
    result, err := device.ReadBlock(0)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(result[:21])) // "Hello, Block Storage!"
}
```

### Open an Existing Encrypted Disk

```go
device, err := block.Open("disk.img", "mypassword")
if err != nil {
    if err == block.ErrWrongPassword {
        fmt.Println("Incorrect password!")
    }
    panic(err)
}
defer device.Close()

// Read and write blocks...
```

## API Reference

### Functions

| Function | Description |
|----------|-------------|
| `New(name, sizeInBytes, password)` | Create a new encrypted block device |
| `Open(name, password)` | Open an existing encrypted block device |

### Methods

| Method | Description |
|--------|-------------|
| `WriteBlock(blockNum, data)` | Encrypt and write data to a block |
| `ReadBlock(blockNum)` | Read and decrypt data from a block |
| `BlockSize()` | Returns the block size (default: 4096) |
| `BlockCount()` | Returns the total number of data blocks |
| `Close()` | Close the device |

### Size Constants

```go
block.KB  // 1024 bytes
block.MB  // 1024 * 1024 bytes
block.GB  // 1024 * 1024 * 1024 bytes
```

## How Encryption Works

1. **Key Derivation**: Password + random salt → PBKDF2 (100k iterations) → 32-byte AES key
2. **Block Encryption**: Each block is encrypted with AES-256-CTR using a unique IV derived from the block number
3. **Verification**: A known plaintext is encrypted and stored in the header to verify the password on open

```
Password: "mysecret"
    │
    ▼ PBKDF2 (password + salt, 100,000 iterations)
    │
AES-256 Key: [32 bytes]
    │
    ▼ For each block
    │
IV = blockNumber → AES-CTR → Encrypted Block
```

## Security Notes

- The header (block 0) is **not encrypted** because it contains the salt needed to derive the key
- Each block uses a unique IV derived from its block number
- PBKDF2 with 100,000 iterations makes brute-force attacks slow
- This is an educational project - for production use, consider established solutions like LUKS

## Verify Encryption

After creating a disk, you can verify the data is encrypted:

```bash
# Check magic bytes
head -c 6 disk.img
# Output: BFSENC

# The rest of the file should be unreadable (encrypted)
cat disk.img | head -c 100
# Output: BFSENC[random bytes...]
```

## Project Structure

```
block-storage/
├── cmd/
│   └── bfs/
│       └── main.go          # Example CLI
├── pkg/
│   └── block/
│       ├── doc.go           # Package documentation
│       ├── block.go         # BlockDevice implementation
│       ├── crypto.go        # Encryption functions
│       └── error.go         # Error definitions
├── go.mod
└── README.md
```

## Future Improvements

- [x] Add `Open()` function to reopen existing disks
- [ ] File system layer on top of blocks
- [ ] Block bitmap for tracking free/used blocks
- [ ] Journaling for crash recovery
- [ ] Compression support

## License

MIT

## Author

Ahmed Wanes - Learning block storage and encryption concepts