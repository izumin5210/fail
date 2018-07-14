package rrr

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStackTrace(t *testing.T) {
	var st0 StackTrace
	func() {
		st0 = newStackTrace(0)
	}()

	var st1 StackTrace
	func() {
		func() {
			st1 = newStackTrace(1)
		}()
	}()

	t.Run("offset 0", func(t *testing.T) {
		assert.NotEmpty(t, st0)
		assert.Equal(t, "TestNewStackTrace", st0[0].Func)
		assert.Regexp(t, regexp.MustCompile(`github.com/\w+/rrr/stack_test.go`), st0[0].File)
		assert.NotZero(t, st0[0].Line)
	})

	t.Run("offset n", func(t *testing.T) {
		assert.NotEmpty(t, st1)
		assert.Equal(t, "TestNewStackTrace", st1[0].Func)
		assert.Regexp(t, regexp.MustCompile(`github.com/\w+/rrr/stack_test.go`), st1[0].File)
		assert.NotZero(t, st1[0].Line)
	})
}

func TestFuncname(t *testing.T) {
	tests := map[string]string{
		"":                                   "",
		"runtime.main":                       "main",
		"github.com/izumin5210/rrr.funcname": "funcname",
		"funcname":                           "funcname",
		"io.copyBuffer":                      "copyBuffer",
		"main.(*R).Write":                    "(*R).Write",
	}

	for input, expect := range tests {
		assert.Equal(t, expect, funcname(input))
	}
}

func TestTrimGOPATH(t *testing.T) {
	gopath := "/home/user"
	file := gopath + "/src/pkg/sub/file.go"
	funcName := "pkg/sub.Type.Method"

	assert.Equal(t, "pkg/sub/file.go", trimGOPATH(funcName, file))
}
