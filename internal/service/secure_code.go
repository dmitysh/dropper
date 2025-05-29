package service

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/dmitysh/dropper/internal/pkg/logger"
)

const (
	minSecretCode = 10
	maxSecretCode = 99
)

type SecureCodeService struct {
	dropCode string
	codeMu   sync.Mutex
}

func NewSecureCodeService() *SecureCodeService {
	return &SecureCodeService{}
}

func (s *SecureCodeService) GenerateCode(ctx context.Context) string {
	s.codeMu.Lock()
	defer s.codeMu.Unlock()

	hostID := strings.Split(getOutboundIP(ctx).String(), ".")[3]
	secretCode := strconv.Itoa(rand.Intn(maxSecretCode-minSecretCode+1) + minSecretCode)

	s.dropCode = hostID + secretCode

	return s.dropCode
}

func (s *SecureCodeService) CodeValid(dropCode string) bool {
	s.codeMu.Lock()
	defer s.codeMu.Unlock()

	return s.dropCode == dropCode
}

func getOutboundIP(ctx context.Context) net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Fatal(ctx, err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
