package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
	"time"
)

var (
	IncorrectCodeErr = errors.New("code is incorrect")
	ConnectionErr    = errors.New("can't connect")
)

const (
	pingTimeout = time.Second * 1
)

type getOptions struct {
	path string
}

func NewGetCommand() *cobra.Command {
	var options getOptions

	cmd := &cobra.Command{
		Use:   "get [OPTIONS]",
		Short: "Get shared file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, &options, args)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.path, "path", "p", ".", "Path where to save file")

	return cmd
}

func runGet(_ *cobra.Command, options *getOptions, args []string) error {
	fileGetterService := service.NewGetFileService()

	dropCode, parseCodeErr := strconv.Atoi(args[0])
	if parseCodeErr != nil {
		log.Println("can't parse drop code: ", parseCodeErr)
		return IncorrectCodeErr
	}

	conn, createConnErr := createConn(hostIDFromDropCode(dropCode))
	if createConnErr != nil {
		log.Println(createConnErr)
		return ConnectionErr
	}
	defer conn.Close()

	fileDropClient := filedrop.NewFileDropClient(conn)
	if pingErr := pingServer(fileDropClient); pingErr != nil {
		return IncorrectCodeErr
	}

	fileStream, getFileStreamErr := getFileStream(args[0], fileDropClient)
	if getFileStreamErr != nil {
		log.Println("can't receive file:")
		return getFileStreamErr
	}

	streamReceiver := filedrop.NewStreamReceiver(fileStream)
	receiveFileErr := fileGetterService.ReceiveAndSaveFileByChunks(streamReceiver, options.path)
	if receiveFileErr != nil {
		log.Println("can't receive file:")
		return errors.New(status.Convert(receiveFileErr).Message())
	}
	log.Println("file saved")

	return nil
}

func hostIDFromDropCode(dropCode int) int {
	return dropCode / 100
}

func pingServer(fileDropClient filedrop.FileDropClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	_, pingErr := fileDropClient.Ping(ctx, &empty.Empty{})

	return pingErr
}

func createConn(hostID int) (*grpc.ClientConn, error) {
	var opts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	port := viper.GetInt("SERVER_PORT")
	localNetID := viper.GetString("LOCAL_NET_ID")

	fullAddr := fmt.Sprintf("%s.%d:%d", localNetID, hostID, port)

	conn, dialErr := grpc.Dial(fullAddr, opts...)
	if dialErr != nil {
		return nil, fmt.Errorf("can't create connection: %w", dialErr)
	}

	return conn, nil
}

func getFileStream(dropCode string, fileDropClient filedrop.FileDropClient) (filedrop.FileDrop_GetFileClient, error) {
	md := metadata.New(map[string]string{"drop-code": dropCode})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	fileStream, getFileErr := fileDropClient.GetFile(ctx, &empty.Empty{})
	if getFileErr != nil {
		return nil, getFileErr
	}

	return fileStream, nil
}
