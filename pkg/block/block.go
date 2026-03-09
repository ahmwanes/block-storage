package block

import (
	"fmt"
	"os"
)

// Common sizes for convenience
const (
	DefaultBlockSize = 4096 // 4KB

	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

// BlockDevice is an encrypted virtual disk that stores data in fixed-size blocks.
// Data is encrypted using AES-256-CTR with a key derived from a password using PBKDF2.
//
// Disk layout:
//   - Block 0: Header (magic bytes, salt, verification)
//   - Block 1+: User data (encrypted)
type BlockDevice struct {
	name       string
	file       *os.File
	blockSize  uint32
	blockCount uint64
	key        []byte
}

// New creates a new encrypted block device with the given name, size, and password.
// The password is used to derive an AES-256 encryption key using PBKDF2.
// WARNING: If a file with the same name exists, it will be overwritten.
func New(name string, sizeInBytes uint64, password string) (*BlockDevice, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// Truncate changes the size of the file.
	// It does not change the I/O offset.
	if err := f.Truncate(int64(sizeInBytes)); err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// generate salt
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// derive key from password + salt
	key, err := deriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// create verification (encrypt known value to verify password later)
	verification, err := createVerification(key)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// write header at block 0 (NOT encrypted - contains salt we need to read!)
	header := Header{
		Magic:        MagicBytes,
		Salt:         [32]byte(salt),
		Verification: [32]byte(verification),
	}

	// serialize header to bytes and write to block 0
	headerBytes := serializeHeader(header)
	_, err = f.WriteAt(headerBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateDevice, err)
	}

	// blockcount: subtract 1 because block 0 is header
	blockCount := (sizeInBytes / DefaultBlockSize) - 1
	device := &BlockDevice{
		name:       name,
		file:       f,
		blockSize:  DefaultBlockSize,
		blockCount: blockCount,
		key:        key,
	}
	return device, nil
}

// Open opens an existing encrypted block device and verifies the password.
// Returns ErrWrongPassword if the password is incorrect.
// Returns ErrInvalidDevice if the file is not a valid encrypted block device.
func Open(name string, password string) (*BlockDevice, error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToOpenDevice, err)
	}

	// Get file size to calculate block count
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToOpenDevice, err)
	}

	// Read header from block 0
	headerBytes := make([]byte, DefaultBlockSize)
	_, err = f.ReadAt(headerBytes, 0)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToOpenDevice, err)
	}

	header := deserializeHeader(headerBytes)

	// Verify magic bytes
	if header.Magic != MagicBytes {
		return nil, ErrInvalidDevice
	}

	// Derive key from password + salt
	key, err := deriveKey(password, header.Salt[:])
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToOpenDevice, err)
	}

	// Verify password is correct
	if err := checkVerification(key, header.Verification[:]); err != nil {
		return nil, err
	}

	// blockcount: subtract 1 because block 0 is header
	blockCount := (uint64(stat.Size()) / DefaultBlockSize) - 1
	device := &BlockDevice{
		name:       name,
		file:       f,
		blockSize:  DefaultBlockSize,
		blockCount: blockCount,
		key:        key,
	}
	return device, nil
}

// WriteBlock encrypts and writes data to the specified block number.
// Returns ErrBlockOutOfRange if blockNum exceeds the device's block count.
func (bd *BlockDevice) WriteBlock(blockNum uint64, data []byte) error {
	// Check block number is valid
	if blockNum >= bd.blockCount {
		return ErrBlockOutOfRange
	}

	encryptedBlock, err := encryptBlock(bd.key, blockNum, data)
	if err != nil {
		return fmt.Errorf("%w - %v", ErrFailedToWriteBlock, err)
	}

	_, err = bd.file.WriteAt(encryptedBlock, bd.offset(blockNum))
	if err != nil {
		return fmt.Errorf("%w - %v", ErrFailedToWriteBlock, err)
	}
	return nil
}

// ReadBlock reads and decrypts data from the specified block number.
// Returns a byte slice of size BlockSize (4KB by default).
func (bd *BlockDevice) ReadBlock(blockNum uint64) ([]byte, error) {
	buffer := make([]byte, bd.blockSize)

	_, err := bd.file.ReadAt(buffer, bd.offset(blockNum))
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToReadBlock, err)
	}

	decryptedBlock, err := decryptBlock(bd.key, blockNum, buffer)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToReadBlock, err)
	}
	return decryptedBlock, nil
}

// Close closes the BlockDevice file.
func (bd *BlockDevice) Close() error {
	return bd.file.Close()
}

// BlockSize returns the block size of the BlockDevice.
func (bd *BlockDevice) BlockSize() uint32 {
	return bd.blockSize
}

// BlockCount returns the number of blocks in the BlockDevice.
func (bd *BlockDevice) BlockCount() uint64 {
	return bd.blockCount
}

// Calculate byte offset: user's block 0 = byte 4096 (skip header block)
func (bd *BlockDevice) offset(blockNum uint64) int64 {
	// +1 because block 0 on disk is the header
	return int64(blockNum+1) * int64(bd.blockSize)
}

// serializeHeader converts Header struct to bytes for writing to disk.
func serializeHeader(h Header) []byte {
	buf := make([]byte, DefaultBlockSize)
	copy(buf[0:6], h.Magic[:])
	copy(buf[6:38], h.Salt[:])
	copy(buf[38:70], h.Verification[:])
	return buf
}

// deserializeHeader converts bytes from disk to Header struct.
func deserializeHeader(buf []byte) Header {
	var h Header
	copy(h.Magic[:], buf[0:6])
	copy(h.Salt[:], buf[6:38])
	copy(h.Verification[:], buf[38:70])
	return h
}
