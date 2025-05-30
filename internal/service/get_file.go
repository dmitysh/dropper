package service

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
)

type ChunkReceiver interface {
	Receive() ([]byte, error)
	Meta() (map[string]string, error)
}

type GetFileService struct {
	path          string
	encryptionKey string
}

func NewGetFileService(path string, encryptionKey string) *GetFileService {
	return &GetFileService{
		path:          path,
		encryptionKey: encryptionKey,
	}
}

func (f *GetFileService) ReceiveAndSaveFileByChunks(fileReceiver ChunkReceiver) error {
	md, mdErr := fileReceiver.Meta()
	if mdErr != nil {
		return errors.New("incorrect meta")
	}

	_, ok := md["filename"]
	if !ok {
		return errors.New("no filename in meta")
	}

	var stream cipher.Stream
	if f.encryptionKey != "" {
		ivEncoded, ok := md["iv"]
		if !ok && f.encryptionKey != "" {
			return errors.New("no iv in meta")
		}

		iv, err := base64.StdEncoding.DecodeString(ivEncoded)
		if err != nil {
			return fmt.Errorf("can't decoode iv: %w", err)
		}

		block, err := aes.NewCipher([]byte(f.encryptionKey))
		if err != nil {
			return fmt.Errorf("can't create cypher: %w", err)
		}

		stream = cipher.NewCTR(block, iv)
	}

	filepath := fp.Join(f.path, md["filename"])
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("can't create file: %w", err)
	}

	success := false
	defer func() {
		if success {
			_ = file.Close()
		} else {
			_ = file.Close()
			_ = os.Remove(filepath)
		}
	}()

	for {
		fileChunk, err := fileReceiver.Receive()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("can't receive chunk: %w", err)
		}

		if stream != nil {
			fileChunkDecrypted := make([]byte, len(fileChunk))
			stream.XORKeyStream(fileChunkDecrypted, fileChunk)
			fileChunk = fileChunkDecrypted
		}

		_, err = file.Write(fileChunk)
		if err != nil {
			return fmt.Errorf("can't write chunk: %w", err)
		}
	}

	success = true

	return nil
}
