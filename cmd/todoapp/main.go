package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_pgx_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool/pgx"
	core_http_middleware "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/server"
	users_postgres_repository "github.com/Trykach34rus/Golang-todoapp/internal/features/users/repository/postgres"
	users_service "github.com/Trykach34rus/Golang-todoapp/internal/features/users/service"
	users_transport_http "github.com/Trykach34rus/Golang-todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {

	ctx,cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger,err := core_logger.NewLogger(core_logger.NewLoggerMust());
	if err != nil {
		fmt.Println("failed to init application logger",err)
		os.Exit(1)
	}

	defer logger.Close()

	logger.Debug("initiazling postges connection pool")	
	pool,err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)

	if err != nil {
		logger.Fatal("failed to init postges connection pool",zap.Error(err))
	}

	defer pool.Close()


	logger.Debug("initiazling feature",zap.String("feature","users"))

	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	userService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(userService)

	logger.Debug("initiazling HTTP Server")
	
	httpServer := core_http_server.NewHTTPServer(core_http_server.NewConfigMust(),logger, 
	core_http_middleware.RequestID(),
  core_http_middleware.Logger(logger),
	core_http_middleware.Trace(),
	core_http_middleware.Panic(),
  ) 	
	
	apiVersionRouter1 := core_http_server.NewApiVersionRouter(core_http_server.ApiVersion1)

  // Example of usage apiVersionRouter2 separate Middleware

	// apiVersionRouter2 := core_http_server.NewApiVersionRouter(
	// 	core_http_server.ApiVersion2,
	// 	core_http_middleware.Dummy("api v2 middleware"),
	// )

	apiVersionRouter1.RegisterRoutes(usersTransportHTTP.Routes()...)
	// apiVersionRouter2.RegisterRoutes(usersTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter1)
	
	if err := httpServer.Run(ctx );err != nil {
		logger.Error("HTTP Server run error",zap.Error(err))
	}
}