package service

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	ErrFileAlreadyRequested = errors.New("file already requested")
)

type ChunkSender interface {
	Send(b []byte) error
}

type SendFileService struct {
	once          sync.Once
	fileChunkSize int
}

func NewSendFileService(fileChunkSize int) *SendFileService {
	return &SendFileService{
		fileChunkSize: fileChunkSize,
	}
}

func (f *SendFileService) SendFileByChunks(filepath string, fileSender ChunkSender) error {
	isFirstAttempt := false
	var sendErr error

	f.once.Do(func() {
		if err := f.sendFile(filepath, fileSender); err != nil {
			sendErr = err
		}
		isFirstAttempt = true
	})

	if !isFirstAttempt {
		return ErrFileAlreadyRequested
	}

	return sendErr
}

func (f *SendFileService) sendFile(filepath string, fileSender ChunkSender) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("can't open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, f.fileChunkSize)

	for {
		n, err := reader.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("can't read chunk: %w", err)
		}

		err = fileSender.Send(buf[:n])
		if err != nil {
			return fmt.Errorf("can't send chunk: %w", err)
		}
	}

	return nil
}
