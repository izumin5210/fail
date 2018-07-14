apperrors
=========

[![Build Status](https://travis-ci.org/creasty/apperrors.svg?branch=master)](https://travis-ci.org/creasty/apperrors)
[![codecov](https://codecov.io/gh/creasty/apperrors/branch/master/graph/badge.svg)](https://codecov.io/gh/creasty/apperrors)
[![GoDoc](https://godoc.org/github.com/creasty/apperrors?status.svg)](https://godoc.org/github.com/creasty/apperrors)
[![License](https://img.shields.io/github/license/creasty/apperrors.svg)](./LICENSE)

Better error handling solution especially for application server.

`apperrors` provides contextual metadata to errors.

- Stack trace
- Additional information
- Status code (for a HTTP server)
- Reportability (for an integration with error reporting service)


Why
---

Since `error` type in Golang is just an interface of [`Error()`](https://golang.org/ref/spec#Errors) method, it doesn't have a stack trace at all. And these errors are likely passed from function to function, you cannot be sure where the error occurred in the first place.  
Because of this lack of contextual metadata, debugging is a pain in the ass.

### How different from [pkg/errors](https://github.com/pkg/errors)

:memo: `apperrors` supports `pkg/errors`. It reuses `pkg/errors`'s stack trace data of the innermost (root) error, and converts into `apperrors`'s data type.

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
	return apperrors.New("invalid email address")
}
```

### Example: Creating from an existing error

```go
_, err := ioutil.ReadAll(r)
if err != nil {
	return apperrors.Wrap(err)
}
```


Annotate an error
-----------------

```go
func WithMessage(msg string) Option
```

WithMessage annotates with the message.

```go
func WithStatusCode(code int) Option
```

WithStatusCode annotates with the status code.

```go
func WithReport() Option
```

WithReport annotates with the reportability.

### Example: Adding all contexts

```go
_, err := ioutil.ReadAll(r)
if err != nil {
	return apperrors.Wrap(
		err,
		apperrors.WithMessage("read failed"),
		apperrors.WithStatusCode(http.StatusBadRequest),
		apperrors.WithReport(),
	)
}
```


Extract context from an error
-----------------------------

```go
func Unwrap(err error) *Error
```

Unwrap extracts an underlying \*apperrors.Error from an error.  
If the given error isn't eligible for retriving context from,
it returns nil

```go
type Error struct {
	// Err is the original error (you might call it the root cause)
	Err error
	// Message is an annotated description of the error
	Message string
	// StatusCode is a status code that is desired to be used for a HTTP response
	StatusCode int
	// Report represents whether the error should be reported to administrators
	Report bool
	// StackTrace is a stack trace of the original error
	// from the point where it was created
	StackTrace StackTrace
}
```

### Example

Here's a minimum executable example describing how `apperrors` works.

```go
package main

import (
	"errors"

	"github.com/creasty/apperrors"
	"github.com/k0kubun/pp"
)

func errFunc0() error {
	return errors.New("this is the root cause")
}
func errFunc1() error {
	return apperrors.Wrap(errFunc0())
}
func errFunc2() error {
	return apperrors.Wrap(errFunc1(), apperrors.WithMessage("fucked up!"))
}
func errFunc3() error {
	return apperrors.Wrap(errFunc2(), apperrors.WithStatusCode(500), apperrors.WithReport())
}

func main() {
	err := errFunc3()
	pp.Println(err)
}
```

```sh-session
$ go run main.go
&apperrors.Error{
  Err:        &errors.errorString{s: "this is the root cause"},
  Message:    "fucked up!",
  StatusCode: 500,
  Report:     true,
  StackTrace: apperrors.StackTrace{
    apperrors.Frame{Func: "errFunc1", File: "main.go", Line: 13},
    apperrors.Frame{Func: "errFunc2", File: "main.go", Line: 16},
    apperrors.Frame{Func: "errFunc3", File: "main.go", Line: 19},
    apperrors.Frame{Func: "main", File: "main.go", Line: 23},
    apperrors.Frame{Func: "main", File: "runtime/proc.go", Line: 194},
    apperrors.Frame{Func: "goexit", File: "runtime/asm_amd64.s", Line: 2198},
  },
}
```

### Example: Server-side error reporting with [gin-gonic/gin](https://github.com/gin-gonic/gin)

Prepare a simple middleware and modify to satisfy your needs:

```go
package middleware

import (
	"net/http"

	"github.com/creasty/apperrors"
	"github.com/creasty/gin-contrib/readbody"
	"github.com/gin-gonic/gin"

	// Only for example
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
)

// ReportError handles an error, changes status code based on the error,
// and reports to an external service if necessary
func ReportError(c *gin.Context, err error) {
	appErr := apperrors.Unwrap(err)
	if appErr == nil {
		// As it's a "raw" error, `StackTrace` field left unset.
		// And it should be always reported
		appErr = &apperrors.Error{
			Err:    err,
			Report: true,
		}
	}

	convertAppError(appErr)

	// Send the error to an external service
	if appErr.Report {
		go uploadAppError(c.Copy(), appErr)
	}

	// Expose an error message in the header
	if appErr.Message != "" {
		c.Header("X-App-Error", appErr.Message)
	}

	// Set status code accordingly
	if appErr.StatusCode > 0 {
		c.Status(appErr.StatusCode)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func convertAppError(err *apperrors.Error) {
	// If the error is from ORM and it says "no record found,"
	// override status code to 404
	if err.Err == gorm.ErrRecordNotFound {
		err.StatusCode = http.StatusNotFound
		return
	}
}

func uploadAppError(c *gin.Context, err *apperrors.Error) {
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
