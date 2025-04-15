package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize  = 16
	nonceSize = 12
	keySize   = 32 // AES-256
	iter      = 100_000
)

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, iter, keySize, sha256.New)
}

func encryptFile(password, filePath string) error {
	plain, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	salt := make([]byte, saltSize)
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	out := bytes.Buffer{}
	out.Write(salt)
	out.Write(nonce)
	out.Write(ciphertext)

	outputFile := filePath + ".enc"
	return os.WriteFile(outputFile, out.Bytes(), 0644)
}

func decryptFile(password, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if len(data) < saltSize+nonceSize {
		return errors.New("invalid encrypted file")
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	outputFile := strings.TrimSuffix(filePath, ".enc")
	return os.WriteFile(outputFile, plain, 0644)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go encrypt --password <password> <file>")
		fmt.Println("  go run main.go decrypt --password <password> <file>")
		return
	}

	action := os.Args[1]

	password := flag.String("password", "", "password to encrypt/decrypt")
	flag.CommandLine.Parse(os.Args[2:])

	if *password == "" {
		fmt.Println("Error: password is required")
		return
	}

	if flag.NArg() < 1 {
		fmt.Println("Error: missing file path")
		return
	}

	filePath := flag.Arg(0)

	switch action {
	case "encrypt":
		if err := encryptFile(*password, filePath); err != nil {
			fmt.Println("Encryption failed:", err)
		} else {
			fmt.Println("File encrypted successfully:", filePath+".enc")
		}
	case "decrypt":
		if err := decryptFile(*password, filePath); err != nil {
			fmt.Println("Decryption failed:", err)
		} else {
			fmt.Println("File decrypted successfully:", strings.TrimSuffix(filePath, ".enc"))
		}
	default:
		fmt.Println("Unknown command:", action)
	}
}
