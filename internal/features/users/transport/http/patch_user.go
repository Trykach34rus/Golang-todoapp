package users_transport_http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Trykach34rus/Golang-todoapp/internal/core/domain"
	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_request "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/request"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
	core_http_types "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/types"
)

type PatchUserRequest struct {
	FullName core_http_types.Nullable[string] `json:"full_name" swaggertype:"string" example:"Max Snow"`
	PhoneNumber core_http_types.Nullable[string] `json:"phone_number" swaggertype:"string" example:"+79876543210"`
}

type PatchUserResponse UserDTOResponse


// PatchUser    godoc
// @Summary     Изменить задачу
// @Description Изменение информации об уже существующем пользователе
// @Description ### Логика обновления полей (Three-state logic):
// @Description 1.**Поле не передано**: `phone_number` игнорируется, значение в БД не меняется 
// @Description 2.**Явно передано значение**: `"phone_number":"+79876543210"` - устанавливает номер телефона в бд
// @Description 3.**Передан null**: `"phone_number": null` - очищает поле в БД (set to NULL)
// @Description Ограничения: `full_name` не может быть выставлен как NULL	
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id path int true "ID изменяемого пользователя"
// @Param       request body PatchUserRequest true "PatchUser тело запроса"
// @Success     200 {object} PatchUserResponse"успешно изменённый пользователь"
// @Failure     400 {object} core_http_response.ErrorResponse "BadRequest"
// @Failure     404 {object} core_http_response.ErrorResponse "User not found"
// @Failure     409 {object} core_http_response.ErrorResponse "Conflict"
// @Failure     500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router      /users/{id} [patch]
func (r *PatchUserRequest)Validate() error  {
	if r.FullName.Set {
		if r.FullName.Value == nil {
			return fmt.Errorf("`FullName` can't be NULL")
		}

		fullNameLen := len([]rune(*r.FullName.Value))
		if fullNameLen < 3 || fullNameLen > 100 {
			return fmt.Errorf("`FullName` must be between 3 and 100 symbols")
		}
	}

	if r.PhoneNumber.Set {
		if r.PhoneNumber.Value != nil {
			phoneNumberLen := len([]rune(*r.PhoneNumber.Value))
			if phoneNumberLen < 10 || phoneNumberLen >15 {
				return fmt.Errorf("`PhoneNumber` must between 10 and 15 symbols")
			}
			if !strings.HasPrefix(*r.PhoneNumber.Value,"+") {
				return fmt.Errorf("`PhoneNumber`must start with '+' symbol")
			}
		}
	}
	return nil
}

func (h *UsersHTTPHandler) PatchUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log,rw)

	userID,err := core_http_request.GetIntPathValue(r,"id")

	if err != nil {
		responseHandler.ErrorResponse(
			err,
		"failed to get user id path value",
	 )
	 return
	}
	

	var request PatchUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r,&request); err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to decode and validate HTTP request",
		)
		return
	}

	userPatch := userPatchFromRequest(request)

	userDomain,err := h.usersService.PatchUser(ctx,userID,userPatch)

	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"failed to patch user ",
		)
		return
	}

	response := PatchUserResponse(userDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response,http.StatusOK)

}


func userPatchFromRequest(request PatchUserRequest) domain.UserPatch  {
	return domain.NewUserPatch(
		request.FullName.ToDomain(),
		request.PhoneNumber.ToDomain(),
	)
}