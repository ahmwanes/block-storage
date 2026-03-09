package block

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Header is stored in block 0 of the disk image (unencrypted).
// It contains the information needed to derive the encryption key
// and verify the password is correct.
type Header struct {
	Magic        [6]byte  // "BFSENC" - identifies this as our format
	Salt         [32]byte // Random salt for PBKDF2 key derivation
	Verification [32]byte // Encrypted known value to verify password
}

const (
	SaltSize         = 32
	PBKDF2Iterations = 100000
	VerifyPlaintext  = "BFSVERIFY!"
)

var (
	MagicBytes = [6]byte{'B', 'F', 'S', 'E', 'N', 'C'}
)

// deriveKey derives a key from a password and salt using PBKDF2.
func deriveKey(password string, salt []byte) ([]byte, error) {
	key, err := pbkdf2.Key(sha256.New, password, salt, PBKDF2Iterations, SaltSize)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToCreateKey, err)
	}
	return key, nil
}

// generateSalt generates a cryptographically secure random salt.
// The salt is used with PBKDF2 to derive the encryption key.
func generateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToGenerateSalt, err)
	}
	return salt, nil
}

// encryptBlock encrypts a block of data using AES-CTR
// The blockNum is used to create a unique nonce/IV for each block
func encryptBlock(key []byte, blockNum uint64, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToEncryptBlock, err)
	}

	// Create IV from blockNum (16 bytes, unique per block)
	iv := make([]byte, aes.BlockSize)        // 16 bytes
	binary.BigEndian.PutUint64(iv, blockNum) // Put blockNum in first 8 bytes

	// Create CTR stream - this handles any size!
	stream := cipher.NewCTR(block, iv)

	// Encrypt
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

// decryptBlock decrypts a block of data using AES-CTR.
// The blockNum is used to create the same IV that was used for encryption.
func decryptBlock(key []byte, blockNum uint64, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w - %v", ErrFailedToDecryptBlock, err)
	}

	// Create IV from blockNum (16 bytes, unique per block)
	iv := make([]byte, aes.BlockSize)        // 16 bytes
	binary.BigEndian.PutUint64(iv, blockNum) // Put blockNum in first 8 bytes

	stream := cipher.NewCTR(block, iv)

	// Decrypt
	plainText := make([]byte, len(ciphertext))
	stream.XORKeyStream(plainText, ciphertext)

	return plainText, nil
}

// createVerification encrypts a known plaintext value.
// This is stored in the header and used to verify the password is correct
// when opening an existing disk.
func createVerification(key []byte) ([]byte, error) {
	plaintext := make([]byte, 32)
	copy(plaintext, []byte(VerifyPlaintext))
	return encryptBlock(key, 0, plaintext)
}

// checkVerification decrypts the verification bytes and checks if they
// match the expected plaintext. Returns ErrWrongPassword if they don't match.
func checkVerification(key []byte, verification []byte) error {
	decrypted, err := decryptBlock(key, 0, verification)
	if err != nil {
		return err
	}

	// Check if decrypted value matches our known plaintext
	if string(decrypted[:len(VerifyPlaintext)]) != VerifyPlaintext {
		return ErrWrongPassword
	}
	return nil
}
