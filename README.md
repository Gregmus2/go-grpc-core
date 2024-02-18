## Example of usage

### Standard usage

```go
package main

func main() {
	core.Serve(
		[]core.Server{
			// example of internal server
			{
				Services: []core.Service{
					// provide generated grpc Desc of your service and constructor of implementation
					{ServiceDesc: private.UserService_ServiceDesc, Constructor: presenters.NewPrivate},
				},
				Interceptors: nil,
				Port:         ":9001",
			},
			// example of external server
			{
				Services: []core.Service{
					{ServiceDesc: public.UserService_ServiceDesc, Constructor: presenters.NewPublic},
				},
				Interceptors: []interceptors.Interceptor{
					&interceptors.AuthorizationInterceptor{},
				},
				Port: ":9000",
			},
		},
		// you can provide your config if needed and any other services
		fx.Provide(
			common.NewConfig,
			logic.NewService,
			adapters.NewDB,
			adapters.NewRepository,
		),
	)
}
```
