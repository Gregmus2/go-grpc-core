package interceptors

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Interceptor interface {
	Name() string
	GetConstructor() any
	UnaryInterceptor() grpc.UnaryServerInterceptor
	StreamInterceptor() grpc.StreamServerInterceptor
	DependsOn() []string
}

type Interceptors []Interceptor

func (i *Interceptors) Sort() error {
	var err error
	*i, err = Sort(*i)
	if err != nil {
		return errors.Wrapf(err, "interceptor dependency error")
	}

	return nil
}

func (i *Interceptors) UnaryInterceptorsAsChain() grpc.ServerOption {
	var interceptors []grpc.UnaryServerInterceptor
	for _, interceptor := range *i {
		interceptors = append(interceptors, interceptor.UnaryInterceptor())
	}

	return grpc.ChainUnaryInterceptor(interceptors...)
}

func (i *Interceptors) StreamInterceptorsAsChain() grpc.ServerOption {
	var interceptors []grpc.StreamServerInterceptor
	for _, interceptor := range *i {
		interceptors = append(interceptors, interceptor.StreamInterceptor())
	}

	return grpc.ChainStreamInterceptor(interceptors...)
}
