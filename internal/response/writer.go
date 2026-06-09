package response

import (
	"fmt"
	"io"

	"github.com/bntrtm/http-from-tcp/internal/headers"
)

type writerState int

const (
	StatusInitialized writerState = iota
	StatusWritingHeaders
	StatusWritingBody
	StatusDone
)

type Writer struct {
	io.Writer
	state writerState
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StatusInitialized {
		return fmt.Errorf("could not write status line to response writer; writer not initialized, or status already written")
	}

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

	w.state = StatusWritingHeaders
	return nil
}

func (w *Writer) newLine(closing bool) error {
	subject := "headers"
	if w.state == StatusWritingBody {
		subject = "body"
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		if closing {
			return fmt.Errorf("could not write line-terminating CRLF while writing %s: %w", subject, err)
		}
		return fmt.Errorf("could not write closing CRLF after writing %s: %w", subject, err)
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	switch w.state {
	case StatusInitialized:
		return fmt.Errorf("could not write headers to response writer; status line not written")
	case StatusWritingHeaders:
	case StatusWritingBody:
		return fmt.Errorf("could not write headers to response writer; headers already written")
	case StatusDone:
		return fmt.Errorf("could not write headers to response writer; writer already finished")
	}

	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s", k, v)
		if err != nil {
			return fmt.Errorf("could not write headers for response: %w", err)
		}
		err = w.newLine(false)
		if err != nil {
			return err
		}
	}
	err := w.newLine(true)
	if err != nil {
		return err
	}

	w.state = StatusWritingBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	switch w.state {
	case StatusInitialized:
		return 0, fmt.Errorf("could not write headers to response writer; missing status line, headers")
	case StatusWritingHeaders:
		return 0, fmt.Errorf("could not write body to response writer; headers not written")
	case StatusWritingBody:
	case StatusDone:
		return 0, fmt.Errorf("could not write body to response writer; writer already finished")
	}

	n, err := w.Write(p)
	if err == nil {
		w.state = StatusDone
		return n, nil
	} else {
		return n, fmt.Errorf("error writing body: %w", err)
	}
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	switch w.state {
	case StatusInitialized:
		return 0, fmt.Errorf("could not write headers to response writer; missing status line, headers")
	case StatusWritingHeaders:
		return 0, fmt.Errorf("could not write body to response writer; headers not written")
	case StatusWritingBody:
	case StatusDone:
		return 0, fmt.Errorf("could not write body to response writer; writer already finished")
	}

	written := 0

	chunkLen := fmt.Sprintf("%x", len(p))
	n, err := w.Write([]byte(chunkLen))
	written += n + 2
	if err != nil {
		return n, err
	}
	_ = w.newLine(false)
	written += 2
	n, err = w.Write(p)
	written += n + 2
	if err != nil {
		return n, err
	}
	_ = w.newLine(false)
	written += 2

	return written, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.Write([]byte("0"))
	if err != nil {
		return n, err
	}
	_ = w.newLine(false)
	_ = w.newLine(true)
	return n + 4, nil
}
