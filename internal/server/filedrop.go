package server

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alexsergivan/transliterator"
	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/pathutils"
	"github.com/dmitysh/dropper/internal/pkg/logger"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrIncorrectCode = status.Error(codes.InvalidArgument, "secure code is incorrect")
)

type FileDropServer struct {
	filedrop.UnimplementedFileDropServer
	fileTransferService *service.SendFileService
	codeService         *service.SecureCodeService
	filepath            string

	TransferDone chan struct{}
}

func NewFileDropServer(filepath string, fileTransferService *service.SendFileService, codeService *service.SecureCodeService) *FileDropServer {
	return &FileDropServer{
		fileTransferService: fileTransferService,
		codeService:         codeService,
		filepath:            filepath,
		TransferDone:        make(chan struct{}),
	}
}

func (f *FileDropServer) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *FileDropServer) GetFile(_ *emptypb.Empty, fileStream filedrop.FileDrop_GetFileServer) error {
	ctx := fileStream.Context()
	defer close(f.TransferDone)

	var fullFilepath string
	if pathutils.CheckPathType(f.filepath) == pathutils.Folder {
		fullFilepath = f.filepath + pathutils.ZipArchiveExt
	} else {
		fullFilepath = f.filepath
	}

	trans := transliterator.NewTransliterator(nil)
	err := fileStream.SendHeader(metadata.New(map[string]string{"filename": trans.Transliterate(filepath.Base(fullFilepath), "en")}))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("can't send header: %v", err))
	}

	err = f.checkSecretCode(ctx)
	if err != nil {
		logger.Err(ctx, "file was requested with invalid code")
		return err
	}

	logger.Info(ctx, "file requested")

	streamSender := filedrop.NewStreamSender(fileStream)
	if sendFileErr := f.fileTransferService.SendFileByChunks(fullFilepath, streamSender); sendFileErr != nil {
		return status.Error(codes.Internal, sendFileErr.Error())
	}

	logger.Info(ctx, "file transferred")

	return nil
}

func (f *FileDropServer) checkSecretCode(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "no meta provided")
	}

	dropCodeMeta := md.Get("drop-code")
	if len(dropCodeMeta) != 1 {
		return ErrIncorrectCode
	}

	if !f.codeService.CodeValid(dropCodeMeta[0]) {
		return ErrIncorrectCode
	}

	return nil
}
