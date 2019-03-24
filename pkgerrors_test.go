package fail

import (
	"errors"
	"testing"

	pkgerrors "github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestExtractPkgError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		pkgErr := extractPkgError(nil)
		assert.Nil(t, pkgErr)
	})

	t.Run("not a pkg/errors", func(t *testing.T) {
		pkgErr := extractPkgError(errors.New("error"))
		assert.Nil(t, pkgErr)
	})

	t.Run("pkg/errors.New", func(t *testing.T) {
		err := pkgErrorsNew("message")

		pkgErr := extractPkgError(err)
		assert.NotNil(t, pkgErr)
		assert.Equal(t, []string(nil), pkgErr.Messages)
		assert.Equal(t, err, pkgErr.Err)
		assert.NotEmpty(t, pkgErr.StackTrace)
		assert.Equal(t, "pkgErrorsNew", pkgErr.StackTrace[0].Func)
	})

	t.Run("pkg/errors.Wrap", func(t *testing.T) {
		t.Run("single wrap", func(t *testing.T) {
			err0 := errors.New("error")
			err1 := pkgErrorsWrap(err0, "message")

			pkgErr := extractPkgError(err1)
			assert.NotNil(t, pkgErr)
			assert.Equal(t, []string{"message"}, pkgErr.Messages)
			assert.Equal(t, err0, pkgErr.Err)
			assert.NotEmpty(t, pkgErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", pkgErr.StackTrace[0].Func)
		})

		t.Run("multiple wrap", func(t *testing.T) {
			err0 := errors.New("error")
			err1 := pkgErrorsWrap(err0, "message 1")
			err2 := pkgErrorsWrap(err1, "message 2")

			pkgErr := extractPkgError(err2)
			assert.NotNil(t, pkgErr)
			assert.Equal(t, []string{"message 2", "message 1"}, pkgErr.Messages)
			assert.Equal(t, err0, pkgErr.Err)
			assert.NotEmpty(t, pkgErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", pkgErr.StackTrace[0].Func)
		})

		t.Run("multiple wrap with an empty message (first)", func(t *testing.T) {
			err0 := errors.New("error")
			err1 := pkgErrorsWrap(err0, "")
			err2 := pkgErrorsWrap(err1, "message 2")
			err3 := pkgErrorsWrap(err2, "message 3")

			pkgErr := extractPkgError(err3)
			assert.NotNil(t, pkgErr)
			assert.Equal(t, []string{"message 3", "message 2"}, pkgErr.Messages)
			assert.Equal(t, err0, pkgErr.Err)
			assert.NotEmpty(t, pkgErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", pkgErr.StackTrace[0].Func)
		})

		t.Run("multiple wrap with an empty message (middle)", func(t *testing.T) {
			err0 := errors.New("error")
			err1 := pkgErrorsWrap(err0, "message 1")
			err2 := pkgErrorsWrap(err1, "")
			err3 := pkgErrorsWrap(err2, "message 3")

			pkgErr := extractPkgError(err3)
			assert.NotNil(t, pkgErr)
			assert.Equal(t, []string{"message 3", "message 1"}, pkgErr.Messages)
			assert.Equal(t, err0, pkgErr.Err)
			assert.NotEmpty(t, pkgErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", pkgErr.StackTrace[0].Func)
		})
	})
}

func TestConvertPkgError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		failErr := convertPkgError(nil)
		assert.Nil(t, failErr)
	})

	t.Run("not a pkg/errors", func(t *testing.T) {
		failErr := convertPkgError(errors.New("error"))
		assert.Nil(t, failErr)
	})

	t.Run("wrap", func(t *testing.T) {
		err0 := errors.New("error")
		err1 := pkgErrorsWrap(err0, "message 1")
		err2 := pkgErrorsWrap(err1, "message 2")

		failErr := convertPkgError(err2)
		assert.NotNil(t, failErr)
		assert.Equal(t, []string{"message 2", "message 1"}, failErr.Messages)
		assert.Equal(t, err0, failErr.Err)
		assert.NotEmpty(t, failErr.StackTrace)
		assert.Equal(t, "pkgErrorsWrap", failErr.StackTrace[0].Func)
	})

	t.Run("mixed (inner most)", func(t *testing.T) {
		err0 := New("error")
		err1 := Wrap(err0, WithMessage("message 1"))
		err2 := pkgErrorsWrap(err1, "message 2")

		failErr := convertPkgError(err2)
		assert.NotNil(t, failErr)
		assert.Equal(t, []string{"message 2", "message 1"}, failErr.Messages)
		assert.Equal(t, err0.Error(), failErr.Err.Error())
		assert.NotEmpty(t, failErr.StackTrace)
	})

	t.Run("mixed (middle)", func(t *testing.T) {
		err0 := errors.New("error")
		err1 := pkgErrorsWrap(err0, "message 1")
		err2 := wrapOrigin(err1)
		err3 := pkgErrorsWrap(err2, "message 2")

		failErr := convertPkgError(err3)
		assert.NotNil(t, failErr)
		assert.Equal(t, []string{"message 2", "message 1"}, failErr.Messages)
		assert.Equal(t, err0, failErr.Err)
		assert.NotEmpty(t, failErr.StackTrace)
		assert.Equal(t, "pkgErrorsWrap", failErr.StackTrace[0].Func)
	})
}

func pkgErrorsNew(msg string) error {
	return pkgerrors.New(msg)
}

func pkgErrorsWrap(err error, msg string) error {
	return pkgerrors.Wrap(err, msg)
}
