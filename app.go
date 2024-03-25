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
	"strconv"
	"sync"
)

func Serve(servers []Server, containerProvider fx.Option) {
	// provide services to DI container
	options := []fx.Option{
		provideBasicServices(),
		containerProvider,
		fxLogger(),
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(servers))
	for j, server := range servers {
		protoServices := buildProtoServicesInjection(server)
		interceptors := buildInterceptorsInjection(server)
		module := buildServerModule(j, protoServices, interceptors, server, j)
		options = append(options, module)
	}

	options = append(options, fx.Invoke(wg.Wait))

	fx.New(options...).Run()
}

func buildServerModule(j int, protoServices []interface{}, interceptors []interface{}, server Server, index int) fx.Option {
	return fx.Module(
		strconv.Itoa(j),
		fx.Provide(protoServices...),
		fx.Provide(interceptors...),
		fx.Invoke(buildRunServerFunction(server, index)),
	)
}

func buildRunServerFunction(server Server, index int) any {
	return func(protoServicesIn ServicesIn, interceptorsIn InterceptorsIn, logger *logrus.Entry, config *Config) error {
		grpcServer, err := buildServer(server, protoServicesIn.Services, interceptorsIn.Interceptors)
		if err != nil {
			return errors.Wrapf(err, "error on build server")
		}

		go runServer(logger, config.ListenAddress[index], grpcServer)

		return nil
	}
}

func buildProtoServicesInjection(server Server) []interface{} {
	protoServices := make([]interface{}, 0, len(server.Services))
	for _, service := range server.Services {
		protoServices = append(protoServices,
			fx.Annotate(
				service.Constructor,
				fx.As(new(protoService)),
				fx.ResultTags(`group:"Services"`),
			),
		)
	}
	// to encapsulate protoServices in server module
	protoServices = append(protoServices, fx.Private)

	return protoServices
}

func buildInterceptorsInjection(server Server) []interface{} {
	interceptors := make([]interface{}, 0, len(server.Interceptors))
	for _, interceptor := range append(server.Interceptors, basicInterceptors()...) {
		interceptors = append(interceptors,
			fx.Annotate(
				interceptor.GetConstructor(),
				fx.As(new(i.Interceptor)),
				fx.ResultTags(`group:"Interceptors"`),
			),
		)
	}
	// to encapsulate interceptors in server module
	interceptors = append(interceptors, fx.Private)

	return interceptors
}

func basicInterceptors() []i.Interceptor {
	return []i.Interceptor{
		i.LoggerInterceptor{},
		i.RequestValidationInterceptor{},
		i.ErrorHandlingInterceptor{},
	}
}

func fxLogger() fx.Option {
	return fx.WithLogger(func(logger *logrus.Entry) (fxevent.Logger, error) {
		return &fxlogrus.LogrusLogger{Logger: logger.Logger}, nil
	})
}

func buildServer(server Server, services []protoService, interceptors i.Interceptors) (*grpc.Server, error) {
	err := interceptors.Sort()
	if err != nil {
		return nil, errors.Wrapf(err, "error sorting interceptors")
	}

	serverOptions := make([]grpc.ServerOption, 0)
	serverOptions = append(serverOptions, interceptors.UnaryInterceptorsAsChain())
	if server.Stream {
		serverOptions = append(serverOptions, interceptors.StreamInterceptorsAsChain())
	}
	s := grpc.NewServer(serverOptions...)

	// register proto reflection service
	reflection.Register(s)

	// register proto services
	for j, service := range server.Services {
		s.RegisterService(&service.ServiceDesc, services[j])
	}

	return s, nil
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
