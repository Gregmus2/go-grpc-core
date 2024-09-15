package core

import (
	"github.com/Gregmus2/go-grpc-core/interceptors"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type protoService any

type Server struct {
	Services     []Service
	Interceptors interceptors.Interceptors
	Stream       bool
}

type Service struct {
	Constructor interface{}
	ServiceDesc grpc.ServiceDesc
}

type ServicesIn struct {
	fx.In

	Services []protoService `group:"Services"`
}

type InterceptorsIn struct {
	fx.In

	Interceptors interceptors.Interceptors `group:"Interceptors"`
}
