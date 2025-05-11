package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	// Create a new cipher stream
	cipherStream, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// read the IV from the source, which in this case would be the
	// block size block.BlockSize we read
	iv := make([]byte, cipherStream.BlockSize()) // 16 bytes for AES
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	var (
		buf       = make([]byte, 1024*32)
		stream    = cipher.NewCTR(cipherStream, iv)
		len_total = cipherStream.BlockSize()
	)
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		stream.XORKeyStream(buf[:n], buf[:n])
		len, err := dst.Write(buf[:n])
		if err != nil {
			return 0, err
		}
		len_total += len
	}
	return len_total, nil
}

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	// Create a new cipher stream
	cipherStream, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, cipherStream.BlockSize()) // 16 bytes for AES
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	//prepend the IV to the destination
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}
	var (
		buf       = make([]byte, 1024*32)
		stream    = cipher.NewCTR(cipherStream, iv)
		len_total = cipherStream.BlockSize()
	)
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		stream.XORKeyStream(buf[:n], buf[:n])
		len, err := dst.Write(buf[:n])
		len_total += len
		if err != nil {
			return 0, err
		}
	}
	return len_total, nil
}
