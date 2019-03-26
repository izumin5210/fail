package fail

import (
	"errors"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New("err")
	assert.Equal(t, "err", err.Error())

	failErr := Unwrap(err)
	assert.Equal(t, "err", failErr.Error())
	assert.NotEmpty(t, failErr.StackTrace)
	assert.Equal(t, "TestNew", failErr.StackTrace[0].Func)
}

func TestErrorf(t *testing.T) {
	err := Errorf("err %d", 123)
	assert.Equal(t, "err 123", err.Error())

	failErr := Unwrap(err)
	assert.Equal(t, "err 123", failErr.Error())
	assert.NotEmpty(t, failErr.StackTrace)
	assert.Equal(t, "TestErrorf", failErr.StackTrace[0].Func)
}

func TestError_LastMessage(t *testing.T) {
	err := &Error{
		Err:      errors.New("err"),
		Messages: []string{"message 2", "message 1"},
	}
	assert.Equal(t, "message 2", err.LastMessage())
}

func TestWithMessage(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithMessage("message"))
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithMessage("message"))
		assert.Equal(t, "message: origin", err1.Error())

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, err1.Error(), failErr.Error())
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := &Error{
			Err:      err0,
			Messages: []string{"message 1"},
			Code:     400,
		}
		err2 := Wrap(err1, WithMessage("message 2"))
		assert.Equal(t, "message 2: message 1: origin", err2.Error())

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, err1.Error(), failErr.Error())
			assert.Equal(t, 400, failErr.Code)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, err2.Error(), failErr.Error())
			assert.Equal(t, 400, failErr.Code)
		}
	})
}

func TestWithMessagef(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithMessagef("message %d", 1))
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithMessagef("message %d", 1))
		assert.Equal(t, "message 1: origin", err1.Error())

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, err1.Error(), failErr.Error())
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := &Error{
			Err:      err0,
			Messages: []string{"message 1"},
			Code:     400,
		}
		err2 := Wrap(err1, WithMessagef("message %d", 2))
		assert.Equal(t, "message 2: message 1: origin", err2.Error())

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, err1.Error(), failErr.Error())
			assert.Equal(t, 400, failErr.Code)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, err2.Error(), failErr.Error())
			assert.Equal(t, 400, failErr.Code)
		}
	})
}

func TestWithCode(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithCode(200))
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithCode(200))

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, "origin", failErr.Error())
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := &Error{
			Err:      err0,
			Messages: []string{"message 1"},
			Code:     400,
		}
		err2 := Wrap(err1, WithCode(500))

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, "message 1: origin", failErr.Error())
			assert.Equal(t, 400, failErr.Code)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, "message 1: origin", failErr.Error())
			assert.Equal(t, 500, failErr.Code)
		}
	})
}

func TestWithTags(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithTags("http", "notice_only"))
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithTags("http", "notice_only"))

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, []string{"http", "notice_only"}, failErr.Tags)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithTags("http", "notice_only"))
		err2 := Wrap(err1, WithTags("security"))

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, []string{"http", "notice_only"}, failErr.Tags)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, []string{"http", "notice_only", "security"}, failErr.Tags)
		}
	})
}

func TestWithParams(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithParams(H{"foo": 1, "bar": "baz"}))
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithParams(H{"foo": 1, "bar": "baz"}))

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, H{"foo": 1, "bar": "baz"}, failErr.Params)
	})

	t.Run("short", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithParam("foo", 1))

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, H{"foo": 1}, failErr.Params)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithParams(H{"foo": 1, "bar": "baz"}))
		err2 := Wrap(err1, WithParams(H{"qux": true, "foo": "quux"}))

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, H{"foo": 1, "bar": "baz"}, failErr.Params)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, H{"foo": "quux", "bar": "baz", "qux": true}, failErr.Params)
		}
	})
}

func TestWithIgnorable(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Wrap(nil, WithIgnorable())
		assert.Equal(t, nil, err)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithIgnorable())

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, "origin", failErr.Error())
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := Wrap(err0, WithIgnorable())
		err2 := Wrap(err1, WithIgnorable())

		{
			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, true, failErr.Ignorable)
		}

		{
			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, true, failErr.Ignorable)
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		failErr := Unwrap(nil)
		assert.Nil(t, failErr)
	})
}

func TestWrap(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		failErr := Wrap(nil)
		assert.Nil(t, failErr)
	})

	t.Run("bare", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := wrapOrigin(err0)
		assert.Equal(t, "origin", err1.Error())

		failErr := Unwrap(err1)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, "origin", failErr.Error())
		assert.NotEmpty(t, failErr.StackTrace)
		assert.Equal(t, "wrapOrigin", failErr.StackTrace[0].Func)
	})

	t.Run("already wrapped", func(t *testing.T) {
		err0 := errors.New("origin")

		err1 := wrapOrigin(err0)
		err2 := wrapOrigin(err1)
		assert.Equal(t, "origin", err2.Error())

		failErr := Unwrap(err2)
		assert.Equal(t, err0, failErr.Err)
		assert.Equal(t, "origin", failErr.Error())
		assert.NotEmpty(t, failErr.StackTrace)
		assert.Equal(t, "wrapOrigin", failErr.StackTrace[0].Func)
	})

	t.Run("with pkg/errors", func(t *testing.T) {
		t.Run("pkg/errors.New", func(t *testing.T) {
			err0 := pkgErrorsNew("origin")

			err1 := wrapOrigin(err0)
			assert.Equal(t, "origin", err1.Error())

			failErr := Unwrap(err1)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, "origin", failErr.Error())
			assert.NotEmpty(t, failErr.StackTrace)
			assert.Equal(t, "pkgErrorsNew", failErr.StackTrace[0].Func)
		})

		t.Run("pkg/errors.Wrap", func(t *testing.T) {
			err0 := errors.New("origin")
			err1 := pkgErrorsWrap(err0, "message")

			err2 := wrapOrigin(err1)
			assert.Equal(t, "message: origin", err2.Error())

			failErr := Unwrap(err2)
			assert.Equal(t, err0, failErr.Err)
			assert.Equal(t, "message: origin", failErr.Error())
			assert.NotEmpty(t, failErr.StackTrace)
			assert.Equal(t, "pkgErrorsWrap", failErr.StackTrace[0].Func)
		})
	})
}

func TestAll(t *testing.T) {
	t.Run("e-p-p-f", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3f())
		assert.Equal(t, "2p: 1p: 0e", failErr.Error())
		assert.Equal(t, nil, failErr.Code)
		assert.Equal(t, false, failErr.Ignorable)
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3f",
			"TestAll.func1",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})

	t.Run("e-p-p-f-f", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3f4f())
		assert.Equal(t, "4f: 2p: 1p: 0e", failErr.Error())
		assert.Equal(t, 500, failErr.Code)
		assert.Equal(t, true, failErr.Ignorable)
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3f",
			"errFunc0e1p2p3f4f",
			"TestAll.func2",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})

	t.Run("e-p-p-fg-f", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3fg4f())
		assert.Equal(t, "4f: 2p: 1p: 0e", failErr.Error())
		assert.Equal(t, 500, failErr.Code)
		assert.Equal(t, true, failErr.Ignorable)
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3fg.func1",
			"errFunc0e1p2p3fg4f",
			"TestAll.func3",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})

	t.Run("e-p-p-f-p", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3f4p())
		assert.Equal(t, "4p: 2p: 1p: 0e", failErr.Error())
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3f",
			"errFunc0e1p2p3f4p",
			"TestAll.func4",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})

	t.Run("e-p-p-fg-f-p", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3fg4f5p())
		assert.Equal(t, "5p: 4f: 2p: 1p: 0e", failErr.Error())
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3fg.func1",
			"errFunc0e1p2p3fg4f",
			"errFunc0e1p2p3fg4f5p",
			"TestAll.func5",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})

	t.Run("e-p-p-pg-p", func(t *testing.T) {
		failErr := Unwrap(errFunc0e1p2p3pg4p())
		assert.Equal(t, "4p: 3pg: 2p: 1p: 0e", failErr.Error())
		assert.Equal(t, []string{
			"errFunc0e1p",
			"errFunc0e1p2p",
			"errFunc0e1p2p3pg.func1",
			"errFunc0e1p2p3pg4p",
			"TestAll.func6",
			"tRunner",
		}, funcNamesFromStackTrace(failErr.StackTrace))
	})
}

func wrapOrigin(err error) error {
	return Wrap(err)
}

func funcNamesFromStackTrace(stackTrace StackTrace) (funcNames []string) {
	for _, frame := range stackTrace {
		funcNames = append(funcNames, frame.Func)
	}
	return
}

// Error functions
//
// How to read: `errFunc0e1p2p3fg4f`
//
// Prefix   Error type: e = build-in errors, f = srvc/fail, p = pkg/errors
// |        |
// errFunc 0e 1p 2p 3fg 4f
// ^^^^^^^ |          |
//         Depth      Goroutine involved
//
// errors -> pkg/errors -> pkg/errors -> fail (goroutine) -> fail

func errFunc0e() error {
	return errors.New("0e")
}
func errFunc0e1p() error {
	return pkgerrors.Wrap(errFunc0e(), "1p")
}
func errFunc0e1p2p() error {
	return pkgerrors.Wrap(errFunc0e1p(), "2p")
}
func errFunc0e1p2p3f() error {
	return Wrap(errFunc0e1p2p())
}
func errFunc0e1p2p3f4f() error {
	return Wrap(errFunc0e1p2p3f(), WithMessage("4f"), WithCode(500), WithIgnorable())
}

func errFunc0e1p2p3fg() chan error {
	c := make(chan error)
	go func() {
		c <- Wrap(errFunc0e1p2p())
	}()
	return c
}
func errFunc0e1p2p3fg4f() error {
	return Wrap(<-errFunc0e1p2p3fg(), WithMessage("4f"), WithCode(500), WithIgnorable())
}
func errFunc0e1p2p3fg4f5p() error {
	return pkgerrors.Wrap(errFunc0e1p2p3fg4f(), "5p")
}

func errFunc0e1p2p3pg() chan error {
	c := make(chan error)
	go func() {
		c <- pkgerrors.Wrap(errFunc0e1p2p(), "3pg")
	}()
	return c
}
func errFunc0e1p2p3pg4p() error {
	return pkgerrors.Wrap(<-errFunc0e1p2p3pg(), "4p")
}

func errFunc0e1p2p3f4p() error {
	return pkgerrors.Wrap(errFunc0e1p2p3f(), "4p")
}
