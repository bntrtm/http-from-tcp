package response

import (
	"fmt"
	"io"

	"github.com/bntrtm/http-from-tcp/internal/headers"
)

// crlf (Carriage Return, Line Feed) is the sequence used
// to denote newlines in HTTP requests.
const crlf = "\r\n"

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := ""
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s", statusCode, reasonPhrase)
	if err != nil {
		return fmt.Errorf("could not write status line for response: %w", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers(map[string]string{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	})
}

func writeCRLF(w io.Writer, closing bool) error {
	_, err := w.Write([]byte(crlf))
	if err != nil {
		if closing {
			return fmt.Errorf("could not write CRLF while writing headers: %w", err)
		}
		return fmt.Errorf("could not write closing CRLF after writing headers: %w", err)
	}
	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s", k, v)
		if err != nil {
			return fmt.Errorf("could not write headers for response: %w", err)
		}
		err = writeCRLF(w, false)
		if err != nil {
			return err
		}
	}
	err := writeCRLF(w, true)
	if err != nil {
		return err
	}
	return nil
}
