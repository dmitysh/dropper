package service

import "github.com/dmitysh/dropper/internal/entity"

type SendFile interface {
	SendFileByChunks(filepath string, fileSender ChunkSender) error
}

type GetFile interface {
	ReceiveAndSaveFileByChunks(fileReceiver ChunkReceiver, filepath string) error
}

type Archive interface {
	FolderToTempZIPArchive(folderPath string) (string, error)
}

type SecureCode interface {
	GenerateDropCode() string
	CheckDropCode(dropCode string) bool
}

type ChunkSender interface {
	Send(chunk entity.FileChunk) error
}

type ChunkReceiver interface {
	Receive() (entity.FileChunk, error)
	Meta() (map[string]string, error)
}
