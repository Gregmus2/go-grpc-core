package interceptors

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const LoggerInterceptorName = "LoggerInterceptor"

type LoggerInterceptor struct {
	logger *logrus.Entry
}

func (l LoggerInterceptor) Name() string {
	return LoggerInterceptorName
}

func (l LoggerInterceptor) GetConstructor() any {
	return func(logger *logrus.Entry) (*LoggerInterceptor, error) {
		return &LoggerInterceptor{
			logger: logger,
		}, nil
	}
}

func (l LoggerInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(interceptorLogger(l.logger), buildLoggingOptions()...)
}

func (l LoggerInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(interceptorLogger(l.logger), buildLoggingOptions()...)
}

func (l LoggerInterceptor) DependsOn() []string {
	return []string{}
}

func interceptorLogger(l logrus.FieldLogger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make(map[string]any, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f[k] = v
		}
		l := l.WithFields(f)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func buildLoggingOptions() []logging.Option {
	return []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall, logging.PayloadReceived, logging.PayloadSent),
	}
}
