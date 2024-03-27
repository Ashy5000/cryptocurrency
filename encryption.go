package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/dsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func IsKeyEncrypted() bool {
	// Check if key.json is encrypted.
	contents, err := os.ReadFile("key.json")
	if err != nil {
		fmt.Println("No key found.")
	}
	var key dsa.PrivateKey
	err = json.Unmarshal(contents, &key)
	if err != nil {
		return true
	}
	return false
}

func EncryptKey(password string) {
	plaintext, err := os.ReadFile("key.json")
	if err != nil {
		fmt.Println("No key found.")
	}
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		fmt.Println("Error creating cipher.")
		fmt.Println("Ensure that the password is a multiple of 16 characters long.")
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	cipherText := gcm.Seal(nonce, nonce, plaintext, nil)
	err = os.WriteFile("key.json", cipherText, 0644)
	if err != nil {
		panic(err)
	}
}

func DecryptKey(password string) {
	ciphertext, err := os.ReadFile("key.json")
	if err != nil {
		panic(err)
	}
	block, err := aes.NewCipher([]byte(password))
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("key.json", plaintext, 0644)
	if err != nil {
		panic(err)
	}
}