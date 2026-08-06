package core_http_response

type ErrorResponse struct {
	Error   string `json:"error" example:"error text"`
	Message string `json:"message" example:"readable error message"`
}