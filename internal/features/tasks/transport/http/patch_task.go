package tasks_transport_http

import (
	"fmt"
	"net/http"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_request "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
	core_http_types "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/types"
)

type PatchTaskRequest struct {
	Title core_http_types.Nullable[string] `json:"title" swaggertype:"string" example:"Сходить в зал"`
	Description core_http_types.Nullable[string] `json:"description" swaggertype:"string" example:"Очень подробное описание что нужно по тренировать"`
	Completed core_http_types.Nullable[bool] `json:"completed" swaggertype:"boolean" example:"true"`
} 

type PatchTaskResponse TaskDTOResponse

// PatchTask    godoc
// @Summary     Изменение задачи
// @Description Изменение информации об уже существующей задачи
// @Description ### Логика обновления полей (Three-state logic):
// @Description 1.**Поле не передано**: `description` игнорируется, значение в БД не меняется 
// @Description 2.**Явно передано значение**: `"description":"После работы в 19:00 надо потренить бицуху"` - устанавливает описание в бд
// @Description 3.**Передан null**: `"description": null` - очищает поле в БД (set to NULL)
// @Description Ограничения: `title` и `completed` не может быть выставлен как NULL	
// @Tags        tasks
// @Accept      json
// @Produce     json
// @Param       id path int true "ID изменяемой задачи"
// @Param       request body PatchTaskRequest true "PatchTask тело запроса"
// @Success     200 {object} PatchTaskResponse"успешное изменение задачи"
// @Failure     400 {object} core_http_response.ErrorResponse "BadRequest"
// @Failure     404 {object} core_http_response.ErrorResponse "Author not found"
// @Failure     409 {object} core_http_response.ErrorResponse "Conflict"
// @Failure     500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router      /tasks/{id} [patch]
func (r *PatchTaskRequest)Validate()error  {
	if r.Title.Set {
		if r.Title.Value == nil {
			return fmt.Errorf("`Title`can't be NULL")
		}

		titleLen := len([]rune(*r.Title.Value))
		if titleLen < 1 || titleLen > 100 {
			return fmt.Errorf("`Title` must be between 1 and 100 symbols")
		}
	}

	if r.Description.Set {
		if r.Description.Value != nil  {
			descriptionLen := len([]rune(*r.Description.Value))
			if descriptionLen < 1 || descriptionLen > 1000 {
				return fmt.Errorf("`Description` must be between 1 and 1000 symbols")
			}
		}
	}

	if r.Completed.Set {
		if r.Completed.Value == nil {
			return fmt.Errorf("`Completed` can't be NULL")
		}
	}

	return nil
}

func (h *TasksHTTPHandler) PatchTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log,rw)

	taskID,err := core_http_request.GetIntPathValue(r,"id")

	if err != nil {
	 responseHandler.ErrorResponse(
		err,
		"failed to get taskID to path value",
	)
		return
	}

	var request PatchTaskRequest

	if err := core_http_request.DecodeAndValidateRequest(r,&request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)

		return
	}

	taskPatch := taskPatchFromRequest(request)

	taskDomain,err := h.tasksService.PatchTask(ctx,taskID,taskPatch)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch task",
		)

		return
	}

	response := PatchTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response,http.StatusOK)

}

func taskPatchFromRequest(request PatchTaskRequest) domain.TaskPatch  {
	return domain.NewTaskPatch(
		request.Title.ToDomain(),
		request.Description.ToDomain(),
		request.Completed.ToDomain(),
	)
	
}