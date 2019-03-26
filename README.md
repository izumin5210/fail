fail
====

[![Build Status](https://travis-ci.com/srvc/fail.svg?branch=master)](https://travis-ci.com/srvc/fail)
[![codecov](https://codecov.io/gh/srvc/fail/branch/master/graph/badge.svg)](https://codecov.io/gh/srvc/fail)
[![GoDoc](https://godoc.org/github.com/srvc/fail?status.svg)](https://godoc.org/github.com/srvc/fail)
[![Go project version](https://badge.fury.io/go/github.com%2Fsrvc%2Ffail.svg)](https://badge.fury.io/go/github.com%2Fsrvc%2Ffail)
[![Go Report Card](https://goreportcard.com/badge/github.com/srvc/fail)](https://goreportcard.com/report/github.com/srvc/fail)
[![License](https://img.shields.io/github/license/srvc/fail.svg)](./LICENSE)

Better error handling solution especially for application servers.

`fail` provides contextual metadata to errors.

- Stack trace
- Error code (to express HTTP/gRPC status code)
- Reportability (to integrate with error reporting services)
- Additional information (tags and params)


Why
---

Since `error` type in Golang is just an interface of [`Error()`](https://golang.org/ref/spec#Errors) method, it doesn't have a stack trace at all. And these errors are likely passed from function to function, you cannot be sure where the error occurred in the first place.  
Because of this lack of contextual metadata, debugging is a pain in the ass.

<!--
### How different from [pkg/errors](https://github.com/pkg/errors)

:memo: `fail` supports `pkg/errors`. It reuses `pkg/errors`'s stack trace data of the innermost (root) error, and converts into `fail`'s data type.
-->


Create an error
---------------

```go
func New(str string) error
```

New returns an error that formats as the given text.
It also records the stack trace at the point it was called.

```go
func Errorf(format string, args ...interface{}) error
```

Errorf formats according to a format specifier and returns the string
as a value that satisfies error.  
It also records the stack trace at the point it was called.

```go
func Wrap(err error, annotators ...Annotator) error
```

Wrap returns an error annotated with a stack trace from the point it was called,
and with the specified options.  
It returns nil if err is nil.

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
func WithMessage(msg string) Annotator
```

WithMessage annotates an error with the message.

```go
func WithMessagef(msg string, args ...interface{}) Annotator
```

WithMessagef annotates an error with the formatted message.

```go
func WithCode(code interface{}) Annotator
```

WithCode annotates an error with the code.

```go
func WithIgnorable() Annotator
```

WithIgnorable annotates an error with the reportability.

```go
func WithTags(tags ...string) Annotator
```

WithTags annotates an error with tags.

```go
func WithParam(key string, value interface{}) Annotator
```

WithParam annotates an error with a key-value pair.

```go
// H represents a JSON-like key-value object.
type H map[string]interface{}

func WithParams(h H) Annotator
```

WithParams annotates an error with key-value pairs.


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
// Error is an error that has contextual metadata
type Error struct {
	// Err is the original error (you might call it the root cause)
	Err error
	// Messages is an annotated description of the error
	Messages []string
	// Code is a status code that is desired to be contained in responses, such as HTTP Status code.
	Code interface{}
	// Ignorable represents whether the error should be reported to administrators
	Ignorable bool
	// Tags represents tags of the error which is classified errors.
	Tags []string
	// Params is an annotated parameters of the error.
	Params H
	// StackTrace is a stack trace of the original error
	// from the point where it was created
	StackTrace StackTrace
}
```

### Example

Here's a minimum executable example illustrating how `fail` works.

```go
package main

import (
	"errors"

	"github.com/k0kubun/pp"
	"github.com/srvc/fail"
)

var myErr = fail.New("this is the root cause")

//-----------------------------------------------
type example1 struct{}

func (e example1) func0() error {
	return errors.New("error from third party")
}
func (e example1) func1() error {
	return fail.Wrap(e.func0())
}
func (e example1) func2() error {
	return fail.Wrap(e.func1(), fail.WithMessage("fucked up!"))
}
func (e example1) func3() error {
	return fail.Wrap(e.func2(), fail.WithCode(500), fail.WithIgnorable())
}

//-----------------------------------------------
type example2 struct{}

func (e example2) func0() error {
	return fail.Wrap(myErr)
}
func (e example2) func1() chan error {
	c := make(chan error)
	go func() {
		c <- fail.Wrap(e.func0(), fail.WithTags("async"))
	}()
	return c
}
func (e example2) func2() error {
	return fail.Wrap(<-e.func1(), fail.WithParam("key", 1))
}
func (e example2) func3() chan error {
	c := make(chan error)
	go func() {
		c <- fail.Wrap(e.func2())
	}()
	return c
}

//-----------------------------------------------
func main() {
	{
		err := (example1{}).func3()
		pp.Println(err)
	}

	{
		err := <-(example2{}).func3()
		pp.Println(err)
	}
}
```

```go
&fail.Error{
	Err: &errors.errorString{s: "error from third party"},
	Messages: []string{"fucked up!"},
	Code:       500,
	Ignorable:  true,
	Tags:       []string{},
	Params:     fail.H{},
	StackTrace: fail.StackTrace{
		fail.Frame{Func: "example1.func1", File: "stack/main.go", Line: 20},
		fail.Frame{Func: "example1.func2", File: "stack/main.go", Line: 23},
		fail.Frame{Func: "example1.func3", File: "stack/main.go", Line: 26},
		fail.Frame{Func: "main", File: "stack/main.go", Line: 58},
	},
}
&fail.Error{
	Err: &errors.errorString{s: "this is the root cause"},
	Messages:   []string{},
	Code:       nil,
	Ignorable:  false,
	Tags:       []string{"async"},
	Params:     {"key": 1},
	StackTrace: fail.StackTrace{
		fail.Frame{Func: "init", File: "stack/main.go", Line: 10},
		fail.Frame{Func: "example2.func0", File: "stack/main.go", Line: 34},
		fail.Frame{Func: "example2.func1.func1", File: "stack/main.go", Line: 39},
		fail.Frame{Func: "example2.func2", File: "stack/main.go", Line: 44},
		fail.Frame{Func: "example2.func3.func1", File: "stack/main.go", Line: 49},
		fail.Frame{Func: "main", File: "stack/main.go", Line: 64},
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
	failErr := fail.Unwrap(err)
	if failErr == nil {
		// As it's a "raw" error, `StackTrace` field left unset.
		// And it should be always reported
		failErr = &fail.Error{
			Err: err,
		}
	}

	convertFailError(failErr)

	// Send the error to an external service
	if !failErr.Ignorable {
		go uploadFailError(c.Copy(), failErr)
	}

	// Expose an error message in the header
	if msg := failErr.LastMessage(); msg != "" {
		c.Header("X-App-Error", msg)
	}

	// Set status code accordingly
	switch code := failErr.Code.(type) {
	case int:
		c.Status(code)
	default:
		c.Status(http.StatusInternalServerError)
	}
}

func convertFailError(err *fail.Error) {
	// If the error is from ORM and it says "no record found,"
	// override status code to 404
	if err.Err == gorm.ErrRecordNotFound {
		err.Code = http.StatusNotFound
		return
	}
}

func uploadFailError(c *gin.Context, err *fail.Error) {
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
