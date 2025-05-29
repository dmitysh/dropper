package grpcutils

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type GRPCServerConfig struct {
	Host string
	Port int
}

func RunAndShutdownServer(serverCfg GRPCServerConfig, grpcServer *grpc.Server, doneCh <-chan struct{}) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serverCfg.Host, serverCfg.Port))
	if err != nil {
		return fmt.Errorf("can't listen: %w", err)
	}
	defer listener.Close()

	go func() {
		<-doneCh
		grpcServer.GracefulStop()
	}()

	if serveErr := grpcServer.Serve(listener); serveErr != nil {
		return serveErr
	}

	return nil
}
