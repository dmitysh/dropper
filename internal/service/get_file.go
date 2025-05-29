package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
)

var (
	IncorrectMetaErr = errors.New("incorrect meta")
)

type ChunkReceiver interface {
	Receive() ([]byte, error)
	Meta() (map[string]string, error)
}

type GetFileService struct{}

func NewGetFileService() *GetFileService {
	return &GetFileService{}
}

func (f *GetFileService) ReceiveAndSaveFileByChunks(fileReceiver ChunkReceiver, path string) error {
	md, err := checkAndGetMeta(fileReceiver)
	if err != nil {
		return err
	}

	filepath := fp.Join(path, md["filename"])
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

		_, err = file.Write(fileChunk)
		if err != nil {
			return fmt.Errorf("can't write chunk: %w", err)
		}
	}

	success = true

	return nil
}

func checkAndGetMeta(fileReceiver ChunkReceiver) (map[string]string, error) {
	md, mdErr := fileReceiver.Meta()
	if mdErr != nil {
		return nil, IncorrectMetaErr
	}

	_, ok := md["filename"]
	if !ok {
		return nil, errors.New("no filename in meta")
	}

	return md, nil
}
