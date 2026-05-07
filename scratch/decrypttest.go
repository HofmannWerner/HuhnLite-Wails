package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func decryptAES(encrypted string, keyString string) (string, error) {
	key := make([]byte, 32)
	copy(key, []byte(keyString))

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func main() {
	pass := "wOvDZk9Nc4PgvlTugzwfad0cUTwZYEH63rJPsvg="
	decrypted, err := decryptAES(pass, "HuhnLite")
	fmt.Println("Decrypted: ", decrypted, err)
}
