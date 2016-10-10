package main

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	pb "lib/proto/auth"
)

const (
	SERVICE = "[CHAT]"
)

var (
	ErrorMethodNotSupported = errors.New("method not supported")

	uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

type server struct {
}

func (s *server) init() {

}

func (s *server) Auth(ctx context.Context, cert *pb.Auth_Certificate) (*pb.Auth_Result, error) {
	switch cert.Type {
	case pb.Auth_UUID:
		if uuidRegexp.MatchString(strings.ToLower(string(cert.Proof))) {
			return &pb.Auth_Result{OK: true, UserId: -1, Body: nil}, nil
		}
		return &pb.Auth_Result{OK: true, UserId: -1, Body: nil}, nil
	case pb.Auth_PLAIN:
	case pb.Auth_TOKEN:
	case pb.Auth_FACEBOOK:
	default:
		return nil, ErrorMethodNotSupported
	}
	return nil, ErrorMethodNotSupported
}
