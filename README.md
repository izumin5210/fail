fail
=========

[![Build Status](https://travis-ci.com/srvc/fail.svg?branch=master)](https://travis-ci.com/srvc/fail)
[![codecov](https://codecov.io/gh/srvc/fail/branch/master/graph/badge.svg)](https://codecov.io/gh/srvc/fail)
[![GoDoc](https://godoc.org/github.com/srvc/fail?status.svg)](https://godoc.org/github.com/srvc/fail)
[![Go project version](https://badge.fury.io/go/github.com%2Fsrvc%2Ffail.svg)](https://badge.fury.io/go/github.com%2Fsrvc%2Ffail)
[![Go Report Card](https://goreportcard.com/badge/github.com/srvc/fail)](https://goreportcard.com/report/github.com/srvc/fail)
[![License](https://img.shields.io/github/license/srvc/fail.svg)](./LICENSE)

Better error handling solution especially for application server.

`fail` provides contextual metadata to errors.

- Stack trace
- Additional information
- Error code (for mapping HTTP status code, gRPC status code, etc.)
- Reportability (for an integration with error reporting service)


This package was forked from [`creasty/apperrors`](https://github.com/srvc/fail).

Why
---

Since `error` type in Golang is just an interface of [`Error()`](https://golang.org/ref/spec#Errors) method, it doesn't have a stack trace at all. And these errors are likely passed from function to function, you cannot be sure where the error occurred in the first place.  
Because of this lack of contextual metadata, debugging is a pain in the ass.

### How different from [pkg/errors](https://github.com/pkg/errors)

:memo: `fail` supports `pkg/errors`. It reuses `pkg/errors`'s stack trace data of the innermost (root) error, and converts into `fail`'s data type.

TBA



Create an error
---------------

```go
func New(str string) error
```

New returns an error that formats as the given text.  
It also annotates the error with a stack trace from the point it was called

```go
func Errorf(format string, args ...interface{}) error
```

Errorf formats according to a format specifier and returns the string
as a value that satisfies error.  
It also annotates the error with a stack trace from the point it was called

```go
func Wrap(err error) error
```

Wrap returns an error annotated with a stack trace from the point it was called.  
It returns nil if err is nil

### Example: Creating a new error

```go
ok := emailRegexp.MatchString("invalid#email.addr")
if !ok {
	return fail.New("invalid email address")
}
```

### Example: Creating from an existing error

```go
_, err := ioutil.ReadAll(r)
if err != nil {
	return fail.Wrap(err)
}
```


Annotate an error
-----------------

```go
func WithMessage(msg string) Option
```

WithMessage annotates with the message.

```go
func WithCode(code interface{}) Option
```

WithCode annotates with the status code.

```go
func WithIgnorable() Option
```

WithIgnorable annotates with the reportability.

```go
func WithTags(tags ...string) Option
```

WithTags annotates with tags.

```go
func WithParam(key string, value interface{}) Option
func WithParams(h H) Option
```

WithParam(s) annotates with key-value pairs.

### Example: Adding all contexts

```go
_, err := ioutil.ReadAll(r)
if err != nil {
	return fail.Wrap(
		err,
		fail.WithMessage("read failed"),
		fail.WithCode(http.StatusBadRequest),
		fail.WithIgnorable(),
	)
}
```


Extract context from an error
-----------------------------

```go
func Unwrap(err error) *Error
```

Unwrap extracts an underlying \*fail.Error from an error.  
If the given error isn't eligible for retriving context from,
it returns nil

```go
type Error struct {
	// Err is the original error (you might call it the root cause)
	Err error
	// Message is an annotated description of the error
	Message string
	// Code is a status code that is desired to be used for a HTTP response
	Code int
	// Ignorable represents whether the error should be reported to administrators
	Ignorable bool
	// StackTrace is a stack trace of the original error
	// from the point where it was created
	StackTrace StackTrace
}
```

### Example

Here's a minimum executable example describing how `fail` works.

```go
package main

import (
	"errors"

	"github.com/srvc/fail"
	"github.com/k0kubun/pp"
)

func errFunc0() error {
	return errors.New("this is the root cause")
}
func errFunc1() error {
	return fail.Wrap(errFunc0())
}
func errFunc2() error {
	return fail.Wrap(errFunc1(), fail.WithMessage("fucked up!"))
}
func errFunc3() error {
	return fail.Wrap(errFunc2(), fail.WithCode(500), fail.WithIgnorable())
}

func main() {
	err := errFunc3()
	pp.Println(err)
}
```

```sh-session
$ go run main.go
&fail.Error{
  Err:        &errors.errorString{s: "this is the root cause"},
  Message:    "fucked up!",
  Code:       500,
  Ignorable:  true,
  StackTrace: fail.StackTrace{
    fail.Frame{Func: "errFunc1", File: "main.go", Line: 13},
    fail.Frame{Func: "errFunc2", File: "main.go", Line: 16},
    fail.Frame{Func: "errFunc3", File: "main.go", Line: 19},
    fail.Frame{Func: "main", File: "main.go", Line: 23},
    fail.Frame{Func: "main", File: "runtime/proc.go", Line: 194},
    fail.Frame{Func: "goexit", File: "runtime/asm_amd64.s", Line: 2198},
  },
}
```

### Example: Server-side error reporting with [gin-gonic/gin](https://github.com/gin-gonic/gin)

Prepare a simple middleware and modify to satisfy your needs:

```go
package middleware

import (
	"net/http"

	"github.com/srvc/fail"
	"github.com/creasty/gin-contrib/readbody"
	"github.com/gin-gonic/gin"

	// Only for example
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
)

// ReportError handles an error, changes status code based on the error,
// and reports to an external service if necessary
func ReportError(c *gin.Context, err error) {
	appErr := fail.Unwrap(err)
	if appErr == nil {
		// As it's a "raw" error, `StackTrace` field left unset.
		// And it should be always reported
		appErr = &fail.Error{
			Err: err,
		}
	}

	convertAppError(appErr)

	// Send the error to an external service
	if !appErr.Ignorable {
		go uploadAppError(c.Copy(), appErr)
	}

	// Expose an error message in the header
	if appErr.Message != "" {
		c.Header("X-App-Error", appErr.Message)
	}

	// Set status code accordingly
	if appErr.Code > 0 {
		c.Status(appErr.Code)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func convertAppError(err *fail.Error) {
	// If the error is from ORM and it says "no record found,"
	// override status code to 404
	if err.Err == gorm.ErrRecordNotFound {
		err.Code = http.StatusNotFound
		return
	}
}

func uploadAppError(c *gin.Context, err *fail.Error) {
	// By using readbody, you can retrive an original request body
	// even when c.Request.Body had been read
	body := readbody.Get(c)

	// Just debug
	pp.Println(string(body[:]))
	pp.Println(err)
}
```

And then you can use like as follows.

```go
r := gin.Default()
r.Use(readbody.Recorder()) // Use github.com/creasty/gin-contrib/readbody

r.GET("/test", func(c *gin.Context) {
	err := doSomethingReallyComplex()
	if err != nil {
		middleware.ReportError(c, err) // Neither `c.AbortWithError` nor `c.Error`
		return
	}

	c.Status(200)
})

r.Run()
```
