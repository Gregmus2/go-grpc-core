package interceptors

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const ErrorHandlingInterceptorName = "ErrorHandlingInterceptor"

var ErrInternal = errors.New("internal server error")

type ErrorMapping map[error]error

type ErrorHandlingInterceptor struct {
	mapping map[error]error
	logger  *logrus.Entry
}

func (i ErrorHandlingInterceptor) GetConstructor() any {
	return func(logger *logrus.Entry, errorMapping ErrorMapping) (*ErrorHandlingInterceptor, error) {
		return &ErrorHandlingInterceptor{
			logger:  logger,
			mapping: errorMapping,
		}, nil
	}
}

func (i ErrorHandlingInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		res, err := handler(ctx, req)
		if err != nil {
			err = i.handleError(err)

			return nil, err
		}

		return res, err
	}
}

func (i ErrorHandlingInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			return i.handleError(err)
		}

		return err
	}
}

func (i ErrorHandlingInterceptor) DependsOn() []string {
	return []string{}
}

func (i ErrorHandlingInterceptor) Name() string {
	return ErrorHandlingInterceptorName
}

func (i ErrorHandlingInterceptor) handleError(err error) error {
	if _, ok := status.FromError(err); ok {
		return err
	}

	for k, v := range i.mapping {
		if errors.Is(err, k) {
			return v
		}
	}

	i.logger.Error(err)

	return ErrInternal
}
