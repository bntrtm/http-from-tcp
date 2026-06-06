package request

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// CRLF (Carriage Return, Line Feed) is the sequence used
// to denote newlines in HTTP requests.
const CRLF = "\r\n"

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
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
func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}

	str := string(data[:idx])
	parts := strings.Split(str, " ")

	if n := len(parts); n < 3 {
		return nil, fmt.Errorf("invalid HTTP request; missing one or more attributes in request-line")
	} else if n > 3 {
		return nil, fmt.Errorf("invalid HTTP request; expected only three attributes in request-line, but got %d", n)
	}
	method, err := validateMethod(parts[0])
	if err != nil {
		return nil, err
	}
	target, err := validateTarget(parts[1])
	if err != nil {
		return nil, err
	}
	version, err := validateVersion(parts[2])
	if err != nil {
		return nil, err
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HTTPVersion:   version,
	}, nil
}

// RequestFromReader returns a new Request object, the number of bytes consumed
// from the input reader, and any relevant error.
func RequestFromReader(reader io.Reader) (*Request, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reqLine, err := parseRequestLine(bytes)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *reqLine,
	}, nil
}
