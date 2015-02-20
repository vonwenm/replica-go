package replica

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HTTPError http status code and error message
type HTTPError struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (r *HTTPError) Error() string {
	return fmt.Sprintf("http error: Code: %d Message: %s", r.Code, r.Message)
}

func newHTTPError(code int, msg []byte) *HTTPError {
	herr := &HTTPError{}
	err := json.Unmarshal(msg, herr)
	if err != nil || herr.Code != code {
		herr.Code = code
		herr.Message = strings.Replace(string(msg), "\n", "", 0)
		return herr
	}
	return herr
}
