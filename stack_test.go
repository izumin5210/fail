package fail

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
		assert.Regexp(t, regexp.MustCompile(`github.com/\w+/fail/stack_test.go`), st0[0].File)
		assert.NotZero(t, st0[0].Line)
	})

	t.Run("offset n", func(t *testing.T) {
		assert.NotEmpty(t, st1)
		assert.Equal(t, "TestNewStackTrace", st1[0].Func)
		assert.Regexp(t, regexp.MustCompile(`github.com/\w+/fail/stack_test.go`), st1[0].File)
		assert.NotZero(t, st1[0].Line)
	})
}

func TestFuncname(t *testing.T) {
	tests := map[string]string{
		"":                              "",
		"runtime.main":                  "main",
		"github.com/srvc/fail.funcname": "funcname",
		"funcname":                      "funcname",
		"io.copyBuffer":                 "copyBuffer",
		"main.(*R).Write":               "(*R).Write",
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

func TestMergeStackTraces(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		inner := StackTrace{}
		outer := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
		}
		result := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
		}

		assert.Equal(t, result, mergeStackTraces(inner, outer))
	})

	t.Run("inner < outer", func(t *testing.T) {
		inner := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
		}
		outer := StackTrace{
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}
		result := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}

		assert.Equal(t, result, mergeStackTraces(inner, outer))
	})

	t.Run("inner > outer (overlapping)", func(t *testing.T) {
		inner := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}
		outer := StackTrace{
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}
		result := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}

		assert.Equal(t, result, mergeStackTraces(inner, outer))
	})

	t.Run("inner > outer (no overlapping frames)", func(t *testing.T) {
		inner := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
		}
		outer := StackTrace{
			{Func: "g2", File: "main.go", Line: 1061},
			{Func: "g3.func1", File: "main.go", Line: 1067},
		}
		result := StackTrace{
			{Func: "init", File: "main.go", Line: 154},
			{Func: "f1", File: "main.go", Line: 157},
			{Func: "f2", File: "main.go", Line: 161},
			{Func: "f3.func1", File: "main.go", Line: 167},
			{Func: "g2", File: "main.go", Line: 1061},
			{Func: "g3.func1", File: "main.go", Line: 1067},
		}

		assert.Equal(t, result, mergeStackTraces(inner, outer))
	})
}
