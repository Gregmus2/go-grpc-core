package interceptors

import (
	"context"
	"github.com/GregmusCo/poll-play-proto-gen/go/private"
	"github.com/caarlos0/env"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const ContextUserIDKey = "user-id"
const UserInterceptorName = "UserInterceptor"

type userConfig struct {
	Address string `env:"USER_SERVICE_ADDRESS" envDefault:"user:9001"`
}

type UserInterceptor struct {
	userClient private.UserServiceClient
}

func (i *UserInterceptor) Init() error {
	cfg := new(userConfig)

	if err := env.Parse(cfg); err != nil {
		return errors.Wrap(err, "error parsing required ENV config for initializing user interceptor")
	}

	cc, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "error creating connection with user service")
	}

	i.userClient = private.NewUserServiceClient(cc)

	return nil
}

func (i *UserInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(i.getUser)
}

func (i *UserInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return auth.StreamServerInterceptor(i.getUser)
}

func (i *UserInterceptor) Name() string {
	return UserInterceptorName
}

func (i *UserInterceptor) DependsOn() []string {
	return []string{
		AuthorizationInterceptorName,
	}
}

func (i *UserInterceptor) getUser(ctx context.Context) (context.Context, error) {
	firebaseID := ctx.Value(ContextFirebaseIdKey).(string)
	if firebaseID == "" {
		panic("firebase id is missing in context")
	}

	resp, err := i.userClient.GetUserByFirebaseID(context.Background(), &private.GetUserByFirebaseIDRequest{
		FirebaseId: firebaseID,
	})

	if err != nil {
		// auth service returns errors with code, so no need to wrap
		return nil, err
	}

	ctx = logging.InjectFields(ctx, logging.Fields{"grpc.request.id", uuid.New().String()})
	ctx = logging.InjectFields(ctx, logging.Fields{"user.user_id", resp.UserId})

	return context.WithValue(ctx, ContextUserIDKey, resp.UserId), nil
}
