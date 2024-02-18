package core

import (
	"github.com/gregmus2/poll-play-golang-core/interceptors"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type protoService any

type Server struct {
	Services     []Service
	Interceptors interceptors.Interceptors
	Stream       bool
	Port         string
}

type Service struct {
	Constructor interface{}
	ServiceDesc grpc.ServiceDesc
}

type ServicesIn struct {
	fx.In

	Services []protoService `group:"Services"`
}
