package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func decrypt(file *os.File, key []byte) error {
	ciphertext, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(ciphertext, ciphertext)
	decryptedFile, err := os.Create(file.Name()[:len(file.Name())-10])
	if err != nil {
		return err
	}
	defer decryptedFile.Close()
	_, err = decryptedFile.Write(ciphertext)
	if err != nil {
		return err
	}
	err = os.Remove(file.Name())
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run decrypt.go [key]")
		os.Exit(1)
	}

	key, err := hex.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	folder := "."
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".encrypted" {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		return decrypt(file, key)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
