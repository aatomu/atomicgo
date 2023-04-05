package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
)

func Encrypt(key, text string) (cipherText string, err error) {
	// strをByteに
	plainBytes := []byte(text)
	hashBytes := []byte(Hash(key, HashSha256))
	keyBytes := hashBytes[:32]

	// AES 暗号化block作成
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return
	}

	// IV を作成
	textBytes := make([]byte, aes.BlockSize+len(plainBytes))
	iv := textBytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		fmt.Printf("err: %s\n", err)
	}

	// Encrypt
	encryptStream := cipher.NewCTR(block, iv)
	encryptStream.XORKeyStream(textBytes[aes.BlockSize:], plainBytes)
	return fmt.Sprintf("%x", textBytes), nil
}

func Decrpt(key, cipherText string) (plainText string, err error) {
	// strをByteに
	hashBytes := []byte(Hash(key, HashSha256))
	keyBytes := hashBytes[:32]
	var cipherBytes []byte
	fmt.Sscanf(cipherText, "%x", &cipherBytes)

	// AES 暗号化block作成
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return
	}

	// Encrypt
	decryptedText := make([]byte, len(cipherBytes[aes.BlockSize:]))
	decryptStream := cipher.NewCTR(block, cipherBytes[:aes.BlockSize])
	decryptStream.XORKeyStream(decryptedText, cipherBytes[aes.BlockSize:])
	return string(decryptedText), nil
}

type hashType uint8

const (
	HashSha1   hashType = 1
	HashSha256 hashType = 2
	HashSha384 hashType = 3
	HashSha512 hashType = 4
)

func Hash(key string, Type hashType) (hash string) {
	// keyをHash化するためにBytesに
	keyBytes := []byte(key)

	switch Type {
	case HashSha1:
		return fmt.Sprintf("%x", sha1.Sum(keyBytes))
	case HashSha256:
		return fmt.Sprintf("%x", sha256.Sum256(keyBytes))
	case HashSha384:
		return fmt.Sprintf("%x", sha512.Sum384(keyBytes))
	case HashSha512:
		return fmt.Sprintf("%x", sha512.Sum512(keyBytes))
	}
	return ""
}
