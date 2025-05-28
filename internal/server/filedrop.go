package server

import (
	"context"
	"fmt"
	"github.com/alexsergivan/transliterator"
	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/pathutils"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	IncorrectCodeErr = status.Error(codes.InvalidArgument, "secure code is incorrect")
)

const maxIncorrectCodesAttempts = 2

type FileDropServer struct {
	filedrop.UnimplementedFileDropServer
	fileTransferService   service.SendFile
	codeService           service.SecureCode
	filepath              string
	StopCh                chan os.Signal
	incorrectCodeAttempts int
	codeMu                sync.Mutex
	incorrectRequests     int
}

func NewFileDropServer(fileTransferService service.SendFile, codeService service.SecureCode,
	filepath string, stopCh chan os.Signal) *FileDropServer {
	return &FileDropServer{
		fileTransferService: fileTransferService,
		StopCh:              stopCh,
		filepath:            filepath,
		codeService:         codeService,
	}
}

func (f *FileDropServer) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *FileDropServer) GetFile(_ *emptypb.Empty, fileStream filedrop.FileDrop_GetFileServer) error {
	var fullFilepath string
	if pathutils.CheckPathType(f.filepath) == pathutils.Folder {
		fullFilepath = f.filepath + ".zip"
	} else {
		fullFilepath = f.filepath
	}

	trans := transliterator.NewTransliterator(nil)
	sendHeaderErr := fileStream.SendHeader(metadata.New(map[string]string{"filename": trans.Transliterate(filepath.Base(fullFilepath), "en")}))
	if sendHeaderErr != nil {
		log.Println(sendHeaderErr)
		return status.Error(codes.Internal, fmt.Sprintf("can't send header: %v", sendHeaderErr))
	}

	if checkSecretCodeErr := f.checkSecretCode(fileStream.Context()); checkSecretCodeErr != nil {
		log.Println(checkSecretCodeErr)
		return checkSecretCodeErr
	}

	log.Println("file requested")

	streamSender := filedrop.NewStreamSender(fileStream)
	if sendFileErr := f.fileTransferService.SendFileByChunks(fullFilepath, streamSender); sendFileErr != nil {
		log.Println(sendFileErr)
		return status.Error(codes.Internal, sendFileErr.Error())
	}

	log.Println("file transferred")
	f.StopCh <- os.Interrupt

	return nil
}

func (f *FileDropServer) checkSecretCode(mdCtx context.Context) error {
	md, ok := metadata.FromIncomingContext(mdCtx)
	if !ok {
		return status.Error(codes.InvalidArgument, "no meta provided")
	}

	dropCodeMeta := md.Get("drop-code")
	if len(dropCodeMeta) != 1 {
		return IncorrectCodeErr
	}

	codeOk := f.codeService.CheckDropCode(dropCodeMeta[0])

	f.codeMu.Lock()
	defer f.codeMu.Unlock()

	if !codeOk {
		f.incorrectCodeAttempts++
		if f.incorrectCodeAttempts > maxIncorrectCodesAttempts && len(f.StopCh) == 0 {
			f.StopCh <- os.Interrupt
			return IncorrectCodeErr
		}

		return IncorrectCodeErr
	}

	return nil
}
