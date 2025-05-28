package grpcutils

import (
	"fmt"
	"github.com/dmitysh/dropper/internal/server"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GRPCServerConfig struct {
	Host string
	Port int
}

func RunAndShutdownServer(serverCfg GRPCServerConfig, grpcServer *grpc.Server, fileDropServer *server.FileDropServer) error {
	listener, listenErr := net.Listen("tcp",
		fmt.Sprintf("%s:%d", serverCfg.Host, serverCfg.Port))
	if listenErr != nil {
		return fmt.Errorf("can't listen: %w", listenErr)
	}
	defer listener.Close()

	signal.Notify(fileDropServer.StopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-fileDropServer.StopCh
		time.Sleep(time.Millisecond * 500)
		log.Println("stop sharing")
		grpcServer.GracefulStop()
	}()

	if serveErr := grpcServer.Serve(listener); serveErr != nil {
		return fmt.Errorf("sharing error: %w", serveErr)
	}

	return nil
}
