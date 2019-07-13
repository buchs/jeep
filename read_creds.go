package read_creds

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var (
	Company   string = "My Company Name"
	Location  string = "My Company Location"
	Filename  string = "creds.dat"
	err       error
	Developer string = "Kevin Buchs"
	Result    string = "No credentials"
)

func ReadCreds() string {
	var fp *os.File
	var buffer bytes.Buffer
	var index int
	var text1 []byte
	var text5 []byte
	Ids := []int{527, 455, 187, 385, 390, 825, 684, 840, 932, 198, 687, 930, 1009, 344, 282, 651, 149, 805, 994, 783, 218, 616, 658, 127, 299, 47, 248, 54, 1015, 821, 888, 475}
	comb := Company + Location + Developer
	text5 = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	for _, id := range Ids {
		index = int(id % len(Ids))
		buffer.WriteString(string(comb[index]))
	}
	text1 = []byte(buffer.String())

	text5, err = ioutil.ReadFile(Filename)

	fp, err = os.Open(Filename)
	if err != nil {
		fmt.Printf("Error opening file: %s, error: %s\n", Filename, err)
		return Result
	}
	index, err = fp.Read(text5)
	if err != nil {
		fmt.Printf("Error reading from file: %s, error: %s\n", Filename, err)
		return Result
	}
	// fmt.Printf("Read %d bytes from file %s\n", index, Filename)
	err = fp.Close()
	if err != nil {
		fmt.Printf("Error closing file: %s, error: %s\n", Filename, err)
		return Result
	}

	result, err := decrypt(text1, text5)
	if err != nil {
		fmt.Printf("Error decrypting creds, error: %s\n", err)
		return Result
	}
	Result = string(result)
	parts := strings.Split(Result, ",")
	Result = parts[0] + "," + parts[1]
	return Result
}

// See alternate IV creation from ciphertext below
//var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	text2 := make([]byte, aes.BlockSize+len(b))
	iv := text2[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(text2[aes.BlockSize:], []byte(b))
	return text2, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
