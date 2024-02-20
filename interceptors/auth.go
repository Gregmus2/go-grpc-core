package interceptors

import (
	"context"
	"github.com/GregmusCo/poll-play-proto-gen/go/private"
	"github.com/caarlos0/env"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const authorizationHeaderName = "authorization"
const ContextFirebaseIdKey = "firebase-id"
const AuthorizationInterceptorName = "AuthorizationInterceptor"

type authenticationConfig struct {
	Address string `env:"USER_SERVICE_ADDRESS" envDefault:"user:9001"`
}

type AuthorizationInterceptor struct {
	userClient private.UserServiceClient
}

func (i *AuthorizationInterceptor) Init() error {
	cfg := new(authenticationConfig)

	if err := env.Parse(cfg); err != nil {
		return errors.Wrap(err, "error parsing required ENV config for initializing authorization interceptor")
	}

	cc, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "error creating connection with authorization service")
	}

	i.userClient = private.NewUserServiceClient(cc)

	return nil
}

func (i *AuthorizationInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(i.auth)
}

func (i *AuthorizationInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return auth.StreamServerInterceptor(i.auth)
}

func (i *AuthorizationInterceptor) Name() string {
	return AuthorizationInterceptorName
}

func (i *AuthorizationInterceptor) DependsOn() []string {
	return []string{}
}

func (i *AuthorizationInterceptor) auth(ctx context.Context) (context.Context, error) {
	token := metadata.ExtractIncoming(ctx).Get(authorizationHeaderName)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization token is required")
	}

	resp, err := i.userClient.Authenticate(context.Background(), &private.AuthenticateRequest{
		Token: token,
	})

	if err != nil {
		// auth service returns errors with code, so no need to wrap
		return nil, err
	}

	ctx = logging.InjectFields(ctx, logging.Fields{"grpc.request.id", uuid.New().String()})
	ctx = logging.InjectFields(ctx, logging.Fields{"user.firebase_id", resp.UserId})

	return context.WithValue(ctx, ContextFirebaseIdKey, resp.UserId), nil
}
