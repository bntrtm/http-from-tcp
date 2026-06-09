package server

import (
	"io"

	"github.com/bntrtm/http-from-tcp/internal/request"
	"github.com/bntrtm/http-from-tcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (e *HandlerError) Write(w io.Writer) {
	_ = response.WriteStatusLine(w, e.StatusCode)
	body := []byte(e.Message)
	_ = response.WriteHeaders(w, response.GetDefaultHeaders(
		len(body),
	))
	_, _ = w.Write(body)
}
