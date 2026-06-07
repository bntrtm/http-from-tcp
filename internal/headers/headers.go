package headers

import (
	"bytes"
	"fmt"
	"regexp"
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

func validateFieldName(s string) (string, error) {
	if strings.Contains(s, " ") {
		return s, fmt.Errorf("bad field-line: whitespace found in field-name")
	}

	keyPattern := `^[!#$%&'*+\-.^_` + "`" + `|~0-9A-Za-z]+$`
	if matched, _ := regexp.MatchString(keyPattern, s); !matched {
		return s, fmt.Errorf("bad field-line: field-name may contain only letters, digits, or special characters !#$%%&'*+-.^_`|~")
	}

	return strings.ToLower(strings.TrimSpace(s)), nil
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

	fieldName, err := validateFieldName(parts[0])
	if err != nil {
		return 0, false, err
	}
	fieldValue := strings.TrimSpace(parts[1])

	if val, ok := h[fieldName]; ok {
		h.Set(fieldName, strings.Join([]string{val, fieldValue}, ", "))
	} else {
		h.Set(fieldName, fieldValue)
	}

	return numParsed, false, nil
}
