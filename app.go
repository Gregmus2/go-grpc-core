package core

import (
	i "github.com/GregmusCo/poll-play-golang-core/interceptors"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	fxlogrus "github.com/takt-corp/fx-logrus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"strconv"
	"sync"
)

func Serve(servers []Server, containerProvider fx.Option) {
	var addresses []string
	if len(os.Args) > 1 {
		addresses = os.Args[1:]
	}

	options := []fx.Option{
		provideBasicServices(),
		containerProvider,
		fxLogger(),
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(servers))
	for j, server := range servers {
		index := j
		localServer := server
		services := make([]interface{}, 0, len(localServer.Services))
		for _, service := range localServer.Services {
			services = append(services,
				fx.Annotate(
					service.Constructor,
					fx.As(new(protoService)),
					fx.ResultTags(`group:"Services"`),
				),
			)
		}
		// to encapsulate services in server module
		services = append(services, fx.Private)

		options = append(options, fx.Module(
			strconv.Itoa(j),
			fx.Provide(services...),
			fx.Invoke(func(services ServicesIn, logger *logrus.Entry, config *Config) error {
				grpcServer, err := buildServer(localServer, logger, services.Services)
				if err != nil {
					return errors.Wrapf(err, "error on build server")
				}

				var address string
				if len(addresses) > index {
					address = addresses[index]
				} else {
					address = config.ListenAddress[index]
				}
				go runServer(logger, address, grpcServer)

				return nil
			}),
		))
	}

	options = append(options, fx.Invoke(wg.Wait))

	fx.New(options...).Run()
}

func fxLogger() fx.Option {
	return fx.WithLogger(func(logger *logrus.Entry) (fxevent.Logger, error) {
		return &fxlogrus.LogrusLogger{Logger: logger.Logger}, nil
	})
}

func buildServer(server Server, logger *logrus.Entry, services []protoService) (*grpc.Server, error) {
	serverOptions := make([]grpc.ServerOption, 0)

	err := server.Interceptors.Prepare()
	if err != nil {
		return nil, errors.Wrapf(err, "error on prepare interceptors")
	}

	unaryInterceptors, err := buildUnaryInterceptors(logger, server.Interceptors)
	if err != nil {
		return nil, errors.Wrapf(err, "error on build unary interceptors")
	}
	serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(unaryInterceptors...))

	if server.Stream {
		streamInterceptors, err := buildStreamInterceptors(logger, server.Interceptors)
		if err != nil {
			return nil, errors.Wrapf(err, "error on build stream interceptors")
		}
		serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	s := grpc.NewServer(serverOptions...)
	reflection.Register(s)

	for j, service := range server.Services {
		s.RegisterService(&service.ServiceDesc, services[j])
	}

	return s, nil
}

func buildUnaryInterceptors(logger *logrus.Entry, interceptors i.Interceptors) ([]grpc.UnaryServerInterceptor, error) {
	basicInterceptors := basicUnaryInterceptors(logger)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(interceptors)+len(basicInterceptors))

	for _, interceptor := range interceptors {
		unaryInterceptors = append(unaryInterceptors, interceptor.UnaryInterceptor())
	}

	return append(unaryInterceptors, basicInterceptors...), nil
}

func buildStreamInterceptors(logger *logrus.Entry, interceptors i.Interceptors) ([]grpc.StreamServerInterceptor, error) {
	basicInterceptors := basicStreamInterceptors(logger)
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(interceptors)+len(basicInterceptors))

	for _, interceptor := range interceptors {
		streamInterceptors = append(streamInterceptors, interceptor.StreamInterceptor())
	}

	return append(streamInterceptors, basicInterceptors...), nil
}

func provideBasicServices() fx.Option {
	return fx.Provide(
		func(config *Config) (logrus.Level, error) {
			return logrus.ParseLevel(config.LogLevel)
		},
		NewLogrusEntry,
		NewConfig,
	)
}

func basicUnaryInterceptors(logger *logrus.Entry) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		i.LogrusUnaryInterceptor(logger),
	}
}

func basicStreamInterceptors(logger *logrus.Entry) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		i.LogrusStreamInterceptor(logger),
	}
}

func runServer(logger *logrus.Entry, address string, server *grpc.Server) {
	logger.Infof("starting listen address %s ...", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		panic(errors.Wrapf(err, "error on listen address %s", address))
	}

	if err := server.Serve(ln); err != nil {
		panic(errors.Wrapf(err, "error on serve address %s", address))
	}
}
