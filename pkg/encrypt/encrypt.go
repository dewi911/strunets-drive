package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
	"io"
)

var secretkey = []byte("salt")

func Encrypt(plaintext string) string {
	block, err := aes.NewCipher(secretkey)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.URLEncoding.EncodeToString(ciphertext)
}

func Decrypt(ciphertext string) string {
	block, err := aes.NewCipher(secretkey)
	if err != nil {
		panic(err)
	}

	decodeCiphertext, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		panic(err)
	}

	if len(decodeCiphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := decodeCiphertext[:aes.BlockSize]
	decodeCiphertext = decodeCiphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(decodeCiphertext, decodeCiphertext)

	return string(decodeCiphertext)
}

func GenerateUUID() string {
	return uuid.New().String()
}
