package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_request "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
)

type GetTaskResponse TaskDTOResponse


// GetTask  godoc
// @Summary     Получить задачу
// @Description Получение задачи по её ID
// @Tags        tasks
// @Produce     json
// @Param       id path int true "ID получаемой задачи"
// @Success     200 {object} GetTaskResponse"успешное получение задачи"
// @Failure     400 {object} core_http_response.ErrorResponse "BadRequest"
// @Failure     404 {object} core_http_response.ErrorResponse "Author not found"
// @Failure     500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [get]
func (h *TasksHTTPHandler) GetTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log,rw)

	taskId, err := core_http_request.GetIntPathValue(r, "id")

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get taskID to path value",
		)
		return
	}

	taskDomain,err := h.tasksService.GetTask(ctx,taskId)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to get task",
		)
		return
	}

	response := GetTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response,http.StatusOK)
}