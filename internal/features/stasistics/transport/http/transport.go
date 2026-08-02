package statistics_tansport_http

import (
	"context"
	"net/http"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_http_server "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/server"
)

type StatisticsHTTPHAndler struct {
	statisticsService StatisticsService
}

type StatisticsService interface {
	GetStatistics(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,		
	)(domain.Statistics,error)
}

func NewStatisticsService(
	statisticsService StatisticsService,
) *StatisticsHTTPHAndler {
	return &StatisticsHTTPHAndler{
		statisticsService: statisticsService,
	}
}

func (h *StatisticsHTTPHAndler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
	{
		Method: http.MethodGet,
		Path: "/statistics",
		Handler: h.GetStatistics,
	},
	}
}
