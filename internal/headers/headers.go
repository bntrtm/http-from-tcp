package headers

import (
	"bytes"
	"fmt"
	"strings"
)

// crlf (Carriage Return, Line Feed) is the sequence used
// to denote newlines in HTTP requests.
const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Set(key, value string) {
	h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	bytes, _, found := bytes.Cut(data, []byte(crlf))
	if !found {
		return 0, false, nil
	}

	numParsed := len(bytes) + len([]byte(crlf))

	if len(bytes) == 0 {
		return numParsed, true, nil
	}

	str := string(bytes)
	parts := strings.SplitN(str, ":", 2)
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("bad field-line: field-name and field-value must be separated with ':'")
	}

	if strings.Contains(parts[0], " ") {
		return 0, false, fmt.Errorf("bad field-line: whitespace found in field-name")
	}
	fieldName := strings.TrimSpace(parts[0])
	fieldValue := strings.TrimSpace(parts[1])

	h.Set(fieldName, fieldValue)

	return numParsed, false, nil
}
