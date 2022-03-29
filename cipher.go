package atomicgo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

func Encrypt(key, text string) (cipherText string, cipherBytes []byte, err error) {
	// strをByteに
	plainBytes := []byte(text)

	// keyをHash化 (強制的にAES-256化)
	hashBytes := []byte(Hash(key, HashSha256))
	keyBytes := hashBytes[:32]

	// AES 暗号化block作成
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return
	}

	// IV を作成
	textBytes := append(hashBytes[:aes.BlockSize], plainBytes...)
	iv := textBytes[:aes.BlockSize]

	// Encrypt
	encryptStream := cipher.NewCTR(block, iv)
	encryptStream.XORKeyStream(textBytes[aes.BlockSize:], plainBytes)
	return fmt.Sprintf("%x", textBytes), textBytes, nil
}

func Decrpt(key, text string, cipherBytes []byte) (cipherText string, err error) {
	// keyをHash化 (強制的にAES-256化)
	hashBytes := []byte(Hash(key, HashSha256))
	keyBytes := hashBytes[:32]

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
