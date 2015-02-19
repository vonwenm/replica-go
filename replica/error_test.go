package replica

import (
	"testing"
)

func TestErrorCover(t *testing.T) {
	err := newHTTPError(404, []byte("not found"))
	msg := "http error: Code: 404 Message: not found"
	if err.Error() != msg {
		t.Errorf("expected '%s', got %s", msg, err.Error())
	}
}
