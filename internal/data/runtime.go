package data

import (
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

var ErrInvalidRuntimeFormat = fmt.Errorf("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error) {
	runtime := fmt.Sprintf("%d mins", r)
	return []byte(strconv.Quote(runtime)), nil
}

func (r *Runtime) UnmarshalJSON(b []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(b))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[0] == "mins" {
		return ErrInvalidRuntimeFormat
	}

	runtime, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = Runtime(runtime)
	return nil
}
