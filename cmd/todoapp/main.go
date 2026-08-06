package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/Trykach34rus/Golang-todoapp/internal/core/config"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_pgx_pool "github.com/Trykach34rus/Golang-todoapp/internal/core/repository/postges/pool/pgx"
	core_http_middleware "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/server"
	postgres_statistics_repository "github.com/Trykach34rus/Golang-todoapp/internal/features/stasistics/repository/postgres"
	service_statistics "github.com/Trykach34rus/Golang-todoapp/internal/features/stasistics/service"
	statistics_tansport_http "github.com/Trykach34rus/Golang-todoapp/internal/features/stasistics/transport/http"
	task_postgres_repository "github.com/Trykach34rus/Golang-todoapp/internal/features/tasks/repository/postgres"
	task_service "github.com/Trykach34rus/Golang-todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/Trykach34rus/Golang-todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/Trykach34rus/Golang-todoapp/internal/features/users/repository/postgres"
	users_service "github.com/Trykach34rus/Golang-todoapp/internal/features/users/service"
	users_transport_http "github.com/Trykach34rus/Golang-todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"

	_ "github.com/Trykach34rus/Golang-todoapp/docs"
)

// @title       Golang Todo API
// @version     1.0
// @description Todo Application REST-API/gRPC schema
// @host        127.0.0.1:5050
// @BasePath    /api/v1
func main() {
	cfg := core_config.NewConfigMust()

	time.Local = cfg.TimeZone

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

	logger.Debug("application time zone",zap.Any("zone",time.Local))

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

	logger.Debug("initiazling feature",zap.String("feature","task"))
	tasksRepository := task_postgres_repository.NewTaskRepository(pool)
	taskService := task_service.NewTaskService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(taskService)

	logger.Debug("initiazling feature",zap.String("feature","statistics"))
	statisticsRepository := postgres_statistics_repository.NewStatisticsRepository(pool)
	statisticService := service_statistics.NewStatisticsService(statisticsRepository)
	statisticTransportHTTP := statistics_tansport_http.NewStatisticsService(statisticService)
 
	logger.Debug("initiazling HTTP Server")
	
	httpServer := core_http_server.NewHTTPServer(core_http_server.NewConfigMust(),logger,
	core_http_middleware.CORS(), 
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

	// apiVersionRouter2.RegisterRoutes(usersTransportHTTP.Routes()...)
	// )

	apiVersionRouter1.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouter1.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouter1.RegisterRoutes(statisticTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter1)
	httpServer.RegisterSwagger()
	
	if err := httpServer.Run(ctx );err != nil {
		logger.Error("HTTP Server run error",zap.Error(err))
	}
}