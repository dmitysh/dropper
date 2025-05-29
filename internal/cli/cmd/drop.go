package cmd

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmitysh/dropper/internal/filedrop"
	"github.com/dmitysh/dropper/internal/pathutils"
	"github.com/dmitysh/dropper/internal/pkg/logger"
	"github.com/dmitysh/dropper/internal/server"
	"github.com/dmitysh/dropper/internal/server/grpcutils"
	"github.com/dmitysh/dropper/internal/service"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	defaultChunkSize = 2 << 20

	serverHost = "0.0.0.0"
	serverPort = 8551
)

var (
	ErrIncorrectPath = errors.New("path to file/folder is not correct")
)

var dropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Share file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		path := args[0]
		var pathToFile string

		switch pathutils.CheckPathType(args[0]) {
		case pathutils.Incorrect:
			logger.Fatal(ctx, ErrIncorrectPath)
		case pathutils.Folder:
			pathToTmpArchive, err := compressFolderToTmpArchive(ctx, path)
			if err != nil {
				logger.Fatalf(ctx, "can't create path to tmp archive: %v", err)
			}
			defer func() {
				err = os.RemoveAll(pathToTmpArchive)
				if err != nil {
					logger.Errf(ctx, "can't remove tmp archive: %v", err.Error())
				}
			}()

			pathToFile = filepath.Join(pathToTmpArchive, filepath.Base(path)+pathutils.ZipArchiveExt)
		case pathutils.File:
			pathToFile = path
		}

		fileSenderService := service.NewSendFileService(defaultChunkSize)
		codeService := service.NewSecureCodeService()
		fileDropServer := server.NewFileDropServer(pathToFile, fileSenderService, codeService)

		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)

		filedrop.RegisterFileDropServer(grpcServer, fileDropServer)

		serverCfg := grpcutils.GRPCServerConfig{
			Host: serverHost,
			Port: serverPort,
		}

		logger.Infof(ctx, "your drop code: %s", codeService.GenerateCode(ctx))

		err := grpcutils.RunAndShutdownServer(serverCfg, grpcServer, fileDropServer.TransferDone)
		if err != nil {
			logger.Fatalf(ctx, "can't serve: %v", err)
		}
	},
}

func compressFolderToTmpArchive(ctx context.Context, path string) (s string, err error) {
	tmpDirPath, err := os.MkdirTemp("", "dropper")
	if err != nil {
		return "", fmt.Errorf("can't create temp directory: %w", err)
	}
	defer func() {
		if err != nil {
			cleanTmpDirErr := os.RemoveAll(tmpDirPath)
			if cleanTmpDirErr != nil {
				logger.Errf(ctx, "can't remove tmp archive: %v", err.Error())
			}
		}
	}()

	archiveFile, err := os.Create(filepath.Join(tmpDirPath, filepath.Base(path)+pathutils.ZipArchiveExt))
	if err != nil {
		return "", fmt.Errorf("can't create archive: %w", err)
	}
	defer archiveFile.Close()

	zw := zip.NewWriter(archiveFile)
	defer zw.Close()

	walker := func(curPath string, info os.FileInfo, recErr error) error {
		if recErr != nil {
			return recErr
		}
		if info.IsDir() {
			return nil
		}

		file, openErr := os.Open(curPath)
		if openErr != nil {
			return openErr
		}
		defer file.Close()

		zipPath, _ := strings.CutPrefix(curPath, path)
		zipPath = strings.Trim(zipPath, "\\./")

		zipFile, createZipErr := zw.Create(zipPath)
		if createZipErr != nil {
			return createZipErr
		}

		_, fileToZipErr := io.Copy(zipFile, file)
		if fileToZipErr != nil {
			return fileToZipErr
		}

		return nil
	}

	err = filepath.Walk(path, walker)
	if err != nil {
		return "", fmt.Errorf("error during recursive archiving: %w", err)
	}

	return tmpDirPath, nil
}
