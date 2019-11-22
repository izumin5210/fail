// +build go1.13

package fail

import (
	"errors"
	"testing"
)

func TestError_Unwrap(t *testing.T) {
	err := errFunc0e()
	failErr := Wrap(err, WithMessage("wrapped"))
	if !errors.Is(failErr, err) {
		t.Errorf("underlying error should be %v", err)
	}
}
