package block

import "errors"

var (
	// BlockStorageErrors
	ErrFailedToCreateDevice = errors.New("failed to create a new block device")
	ErrFailedToOpenDevice   = errors.New("failed to open block device")
	ErrFailedToWriteBlock   = errors.New("failed to write block")
	ErrFailedToReadBlock    = errors.New("failed to read block")
	ErrBlockOutOfRange      = errors.New("block out of range")
	ErrInvalidDevice        = errors.New("invalid device: not a valid BFS encrypted disk")
	ErrFailedToWipeDevice   = errors.New("failed to wipe device")

	// Encryption Errors
	ErrFailedToCreateKey    = errors.New("failed to create encryption key")
	ErrFailedToEncryptBlock = errors.New("failed to encrypt block")
	ErrFailedToDecryptBlock = errors.New("failed to decrypt block")
	ErrFailedToGenerateSalt = errors.New("failed to generate salt")
	ErrWrongPassword        = errors.New("wrong password")
)
