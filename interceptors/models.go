package interceptors

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Interceptor interface {
	Name() string
	Init() error
	UnaryInterceptor() grpc.UnaryServerInterceptor
	StreamInterceptor() grpc.StreamServerInterceptor
	DependsOn() []string
}

type Interceptors []Interceptor

func (i *Interceptors) Prepare() error {
	var err error
	*i, err = Sort(*i)
	if err != nil {
		return errors.Wrapf(err, "interceptor dependency error")
	}

	for _, interceptor := range *i {
		err = interceptor.Init()
		if err != nil {
			return errors.Wrapf(err, "error on init interceptor %s", interceptor.Name())
		}
	}

	return nil
}
