package filedrop

import (
	"github.com/dmitysh/dropper/internal/entity"
)

type StreamSender struct {
	gRPCFileStream FileDrop_GetFileServer
}

func NewStreamSender(fileStream FileDrop_GetFileServer) *StreamSender {
	return &StreamSender{gRPCFileStream: fileStream}
}

func (s *StreamSender) Send(chunk entity.FileChunk) error {
	return s.gRPCFileStream.Send(&FileRequest{ChunkData: chunk})
}

type StreamReceiver struct {
	gRPCFileStream FileDrop_GetFileClient
}

func NewStreamReceiver(fileStream FileDrop_GetFileClient) *StreamReceiver {
	return &StreamReceiver{gRPCFileStream: fileStream}
}

func (s *StreamReceiver) Receive() (entity.FileChunk, error) {
	fileChunk, recvErr := s.gRPCFileStream.Recv()
	if recvErr != nil {
		return entity.FileChunk{}, recvErr
	}

	return fileChunk.GetChunkData(), nil
}

func (s *StreamReceiver) Meta() (map[string]string, error) {
	md, getMdErr := s.gRPCFileStream.Header()
	if getMdErr != nil {
		return nil, getMdErr
	}

	plainMeta := make(map[string]string)

	for k, v := range md {
		plainMeta[k] = v[0]
	}

	return plainMeta, nil
}
