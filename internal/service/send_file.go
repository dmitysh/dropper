package service

import (
	"bufio"
	"crypto/cipher"
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
	ctrStream     cipher.Stream
}

func NewSendFileService(fileChunkSize int, ctrStream cipher.Stream) *SendFileService {
	return &SendFileService{
		fileChunkSize: fileChunkSize,
		ctrStream:     ctrStream,
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

		chunk := buf[:n]
		if f.ctrStream != nil {
			chunk = make([]byte, n)
			f.ctrStream.XORKeyStream(chunk, buf[:n])
		}

		err = fileSender.Send(chunk)
		if err != nil {
			return fmt.Errorf("can't send chunk: %w", err)
		}
	}

	return nil
}
