package statistics_tansport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_request "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TasksCreated              int `json:"tasks_created"`
	TasksCompleted            int `json:"tasks_completed"`
	TasksCompletedRate        *float64 `json:"tasks_completed_rate"`
	TaskAverageCompletionTime *string `json:"tasks_avarage_completion_time"`
}

func (h *StatisticsHTTPHAndler) GetStatistics(rw http.ResponseWriter,r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log,rw)

	userID,from,to,err := getUserIDFromToQueryParams(r)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get userID/from/to query params",
		)
		return
	}

	statistics, err := h.statisticsService.GetStatistics(ctx,userID,from,to)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to statistics",
		)
		return
	}

	response := toDTOFromDomain(statistics)

	responseHandler.JSONResponse(response,http.StatusOK)

}

func toDTOFromDomain(statistics domain.Statistics) GetStatisticsResponse  {
	var avgTime *string
	if statistics.TaskAverageCompletionTime != nil {
		duration := statistics.TaskAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatisticsResponse{
		TasksCreated: statistics.TasksCreated,
		TasksCompleted: statistics.TasksCompleted,
		TasksCompletedRate: statistics.TasksCompletedRate,
		TaskAverageCompletionTime: avgTime,
	}
}



func getUserIDFromToQueryParams(r *http.Request) (*int,*time.Time,*time.Time,error) {

		const (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey = "from"
		toQueryParamKey = "to"

	)

	userID,err := core_http_request.GetIntQueryParam(r,userIDQueryParamKey)

	if err != nil {
		return nil,nil,nil,fmt.Errorf("get 'user _id' query param:%w",err)
	}

	from,err := core_http_request.GetDateQueryParam(r,fromQueryParamKey)

	if err != nil {
		return nil,nil,nil, fmt.Errorf("get 'from' query param:%w",err)
	}

	to,err := core_http_request.GetDateQueryParam(r,toQueryParamKey)

	if err != nil {
		return nil,nil,nil, fmt.Errorf("get 'to' query param:%w",err)
	}

	return userID,from,to,nil

}