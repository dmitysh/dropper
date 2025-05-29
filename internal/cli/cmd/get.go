package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/pkg/logger"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	pingTimeout = time.Second * 3
	localNetID  = "192.168.1"
)

var (
	ErrIncorrectCode = errors.New("code is incorrect")
)

const (
	getFilePathVarName = "path"
)

var (
	getFilePath string
)

func init() {
	getCmd.Flags().StringVarP(&getFilePath, getFilePathVarName, "p", ".", "Path where to save file")
}

var getCmd = &cobra.Command{
	Use:   "get [code]",
	Short: "Get shared file",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		fileGetterService := service.NewGetFileService()

		dropCode, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Fatalf(ctx, "can't parse drop code: %v", err)
		}

		conn, err := createConn(hostIDFromDropCode(dropCode))
		if err != nil {
			logger.Fatalf(ctx, "can't create conn: %v", err)
		}
		defer conn.Close()

		fileDropClient := filedrop.NewFileDropClient(conn)
		err = pingServer(fileDropClient)
		if err != nil {
			logger.Fatalf(ctx, ErrIncorrectCode.Error())
		}

		fileStream, err := getFileStream(ctx, args[0], fileDropClient)
		if err != nil {
			logger.Fatalf(ctx, "can't get file stream: %v", err)
		}

		streamReceiver := filedrop.NewStreamReceiver(fileStream)
		err = fileGetterService.ReceiveAndSaveFileByChunks(streamReceiver, getFilePath)
		if err != nil {
			if status.Code(err) == codes.InvalidArgument {
				logger.Fatal(ctx, "invalid code")
			}
			logger.Fatalf(ctx, "can't receive file: %v", err)
		}

		logger.Infof(ctx, "file saved in directory %s", getFilePath)
	},
}

func hostIDFromDropCode(dropCode int) int {
	return dropCode / 100
}

func pingServer(fileDropClient filedrop.FileDropClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	_, err := fileDropClient.Ping(ctx, &empty.Empty{})
	return err
}

func createConn(hostID int) (*grpc.ClientConn, error) {
	var opts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	fullAddr := fmt.Sprintf("%s.%d:%d", localNetID, hostID, serverPort)

	conn, dialErr := grpc.Dial(fullAddr, opts...)
	if dialErr != nil {
		return nil, fmt.Errorf("can't create connection: %w", dialErr)
	}

	return conn, nil
}

func getFileStream(ctx context.Context, dropCode string, fileDropClient filedrop.FileDropClient) (filedrop.FileDrop_GetFileClient, error) {
	md := metadata.New(map[string]string{"drop-code": dropCode})
	ctx = metadata.NewOutgoingContext(ctx, md)

	fileStream, err := fileDropClient.GetFile(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return fileStream, nil
}
