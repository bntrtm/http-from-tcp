package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// CRLF (Carriage Return, Line Feed) is the sequence used
// to denote newlines in HTTP requests.
const CRLF = "\r\n"

const bufferSize = 8

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type parserState int

const (
	StatusInitialized parserState = iota
	StatusDone
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

func validateMethod(s string) (string, error) {
	asUpper := strings.ToUpper(s)
	if s != asUpper {
		return s, fmt.Errorf("expected method to be in uppercase")
	}
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "TRACE", "CONNECT"}

	if !slices.Contains(validMethods, s) {
		return s, fmt.Errorf("method '%s' was not recognized", s)
	}

	return s, nil
}

func validateTarget(s string) (string, error) {
	if !strings.HasPrefix(s, "/") {
		return s, fmt.Errorf("request-target not written as path")
	}

	return s, nil
}

func validateVersion(s string) (string, error) {
	prefix := `HTTP/`

	if !strings.HasPrefix(s, prefix) {
		return s, fmt.Errorf("HTTP-version missing prefix: %s", prefix)
	}
	version := strings.TrimPrefix(s, prefix)
	if matched, _ := regexp.MatchString(`^\d+\.\d+$`, version); !matched {
		return s, fmt.Errorf("HTTP-version '%s' is not of proper form: DIGIT.DIGIT", version)
	}
	if version != "1.1" {
		return s, fmt.Errorf("incompatible HTTP-version: %s, expected: 1.1", version)
	}

	return version, nil
}

// parseRequestLine builds a new RequestLine given an HTTP request as a string.
func parseRequestLine(data []byte) (*RequestLine, int, error) {
	bytes, _, found := bytes.Cut(data, []byte(CRLF))
	if !found {
		return nil, 0, nil
	}

	numParsed := len(bytes) + len([]byte(CRLF))

	parts := strings.Split(string(bytes), " ")

	if n := len(parts); n < 3 {
		return nil, numParsed, fmt.Errorf("invalid HTTP request; missing one or more attributes in request-line, got only %d", n)
	} else if n > 3 {
		return nil, numParsed, fmt.Errorf("invalid HTTP request; expected only three attributes in request-line, but got %d", n)
	}

	method, err := validateMethod(parts[0])
	if err != nil {
		return nil, numParsed, err
	}
	target, err := validateTarget(parts[1])
	if err != nil {
		return nil, numParsed, err
	}
	version, err := validateVersion(parts[2])
	if err != nil {
		return nil, numParsed, err
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HTTPVersion:   version,
	}, numParsed, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case StatusInitialized:
		l, n, err := parseRequestLine(data)
		if err != nil {
			return n, err
		} else if n == 0 {
			return 0, nil
		}
		r.RequestLine = *l
		r.state = StatusDone

		return n, nil
	case StatusDone:
		return 0, fmt.Errorf("cannot read data with Request in done state")
	default:
		return 0, fmt.Errorf("unknown Request state")
	}
}

// RequestFromReader returns a new Request object, the number of bytes consumed
// from the input reader, and any relevant error.
func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	r := Request{state: StatusInitialized}
	for r.state != StatusDone {
		if len(buf) == cap(buf) {
			grown := make([]byte, len(buf)*2)
			copy(grown, buf)
			buf = grown
		}

		nRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.state = StatusDone
				break
			}
		}
		readToIndex += nRead

		nParsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		size := len(buf) - nParsed
		c := make([]byte, size)
		copy(c, buf[nParsed:])
		buf = c

		readToIndex -= nParsed
	}

	return &r, nil
}
