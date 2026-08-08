package web_transport_http

import (
	core_http_server "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/server"
)

type WebHTTPHandler struct {
	WebService WebService
}

type WebService interface {
	GetMainPage() ([]byte, error)
}

func NewWebHTTPHandler(
	webService WebService,
) *WebHTTPHandler {

	return &WebHTTPHandler{
		WebService: webService,
	}
}

func (h *WebHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Path: "/",
			Handler: h.GetMainPage,
		},
	}
}