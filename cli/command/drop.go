package command

import (
	"errors"
	"fmt"
	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/pathutils"
	"github.com/dmitysh/dropper/internal/server"
	"github.com/dmitysh/dropper/internal/server/grpcutils"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
)

const archiveExt = ".zip"

var (
	IncorrectPathErr = errors.New("path to file/folder is not correct")
	PreparationErr   = errors.New("can't prepare for file transfer")
)

type dropOptions struct {
}

func NewDropCommand() *cobra.Command {
	var options dropOptions

	cmd := &cobra.Command{
		Use:   "drop [OPTIONS]",
		Short: "Share file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDrop(cmd, &options, args)
		},
	}

	return cmd
}

func runDrop(_ *cobra.Command, _ *dropOptions, args []string) error {
	path := args[0]
	var pathToFile string

	switch pathutils.CheckPathType(args[0]) {
	case pathutils.Incorrect:
		return IncorrectPathErr
	case pathutils.Folder:
		pathToTmpArchive, createTmpArchiveErr := createTmpArchive(path)
		if createTmpArchiveErr != nil {
			return createTmpArchiveErr
		}
		defer os.RemoveAll(pathToTmpArchive)

		pathToFile = filepath.Join(pathToTmpArchive, filepath.Base(path)+archiveExt)
	case pathutils.File:
		pathToFile = path
	}

	var fileSenderService service.SendFile
	fileSenderService = service.NewSendFileService(viper.GetInt("CHUNK_SIZE"))

	var codeService service.SecureCode
	codeService = service.NewSecureCodeService()

	fileDropServer := server.NewFileDropServer(fileSenderService,
		codeService,
		pathToFile,
		make(chan os.Signal, 2))

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	filedrop.RegisterFileDropServer(grpcServer, fileDropServer)

	serverCfg := grpcutils.GRPCServerConfig{
		Host: viper.GetString("SERVER_HOST"),
		Port: viper.GetInt("SERVER_PORT"),
	}

	fmt.Println("your drop code:", codeService.GenerateDropCode())

	runSrvErr := grpcutils.RunAndShutdownServer(serverCfg, grpcServer, fileDropServer)
	if runSrvErr != nil {
		log.Println("can't serve:", runSrvErr)
		return PreparationErr
	}

	return nil
}

func createTmpArchive(path string) (string, error) {
	archiveName := filepath.Base(path) + archiveExt

	var archiveService service.Archive
	archiveService = service.NewArchiveService(archiveName)

	pathToTmpArchive, createTmpArchiveErr := archiveService.FolderToTempZIPArchive(path)
	if createTmpArchiveErr != nil {
		log.Println("can't compress folder:", createTmpArchiveErr)
		return "", PreparationErr
	}

	return pathToTmpArchive, nil
}
