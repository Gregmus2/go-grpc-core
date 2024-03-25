package interceptors

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const RequestValidationInterceptorName = "RequestValidationInterceptor"

type Validator interface {
	Validate(request any) error
}

type RequestValidationInterceptor struct {
	logger    *logrus.Entry
	validator Validator
}

func (i RequestValidationInterceptor) Name() string {
	return RequestValidationInterceptorName
}

func (i RequestValidationInterceptor) GetConstructor() any {
	return func(validator Validator, logger *logrus.Entry) (*RequestValidationInterceptor, error) {
		return &RequestValidationInterceptor{
			logger:    logger,
			validator: validator,
		}, nil
	}
}

func (i RequestValidationInterceptor) DependsOn() []string {
	return []string{}
}

func (i RequestValidationInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := i.validator.Validate(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return handler(ctx, req)
	}
}

func (i RequestValidationInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return nil
}
