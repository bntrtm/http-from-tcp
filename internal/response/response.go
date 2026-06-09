package response

import (
	"fmt"
	"net"

	"github.com/bntrtm/http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		Writer: conn,
		state:  StatusInitialized,
	}
}
