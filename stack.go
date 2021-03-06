package fail

import (
	"runtime"
	"strings"
)

const (
	stackMaxSize    = 32
	stackBaseOffset = 3
)

// StackTrace is a stack of Frame from innermost to outermost
type StackTrace []Frame

// Frame represents a single frame of stack trace
type Frame struct {
	Func string
	File string
	Line int64
}

// newFrameFromRuntimeFrame creates Frame from the specified runtime.Frame
func newFrameFromRuntimeFrame(rf runtime.Frame) (f Frame, ok bool) {
	f.Func = funcname(rf.Function)
	f.File = trimGOPATH(rf.Function, rf.File)
	f.Line = int64(rf.Line)

	if strings.HasPrefix(f.File, "runtime/") {
		return
	}

	ok = true
	return
}

// newStackTrace creates StackTrace by callers
func newStackTrace(offset int) StackTrace {
	pcs := make([]uintptr, stackMaxSize)
	n := runtime.Callers(stackBaseOffset+offset, pcs[:])
	return newStackTraceFromPCs(pcs[:n])
}

// newStackTraceFromPCs creates StackTrace from program counters
func newStackTraceFromPCs(pcs []uintptr) (frames StackTrace) {
	runtimeFrames := runtime.CallersFrames(pcs)

	for {
		rf, more := runtimeFrames.Next()
		if frame, ok := newFrameFromRuntimeFrame(rf); ok {
			frames = append(frames, frame)
		}
		if !more {
			break
		}
	}

	return frames
}

// funcname removes the path prefix component of a function's name reported by func.Name().
// Copied from https://github.com/pkg/errors/blob/master/stack.go
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

// Copied from https://github.com/pkg/errors/blob/master/stack.go
func trimGOPATH(name, file string) string {
	// Here we want to get the source file path relative to the compile time
	// GOPATH. As of Go 1.6.x there is no direct way to know the compiled
	// GOPATH at runtime, but we can infer the number of path segments in the
	// GOPATH. We note that fn.Name() returns the function name qualified by
	// the import path, which does not include the GOPATH. Thus we can trim
	// segments from the beginning of the file path until the number of path
	// separators remaining is one more than the number of path separators in
	// the function name. For example, given:
	//
	//    GOPATH     /home/user
	//    file       /home/user/src/pkg/sub/file.go
	//    fn.Name()  pkg/sub.Type.Method
	//
	// We want to produce:
	//
	//    pkg/sub/file.go
	//
	// From this we can easily see that fn.Name() has one less path separator
	// than our desired output. We count separators from the end of the file
	// path until it finds two more than in the function name and then move
	// one character forward to preserve the initial path segment without a
	// leading separator.
	const sep = "/"
	goal := strings.Count(name, sep) + 2
	i := len(file)
	for n := 0; n < goal; n++ {
		i = strings.LastIndex(file[:i], sep)
		if i == -1 {
			// not enough separators found, set i so that the slice expression
			// below leaves file unmodified
			i = -len(sep)
			break
		}
	}
	// get back to 0 or trim the leading separator
	file = file[i+len(sep):]
	return file
}

// mergeStackTraces merges two stack traces
func mergeStackTraces(inner StackTrace, outer StackTrace) StackTrace {
	innerLen := len(inner)
	outerLen := len(outer)

	if innerLen > outerLen {
		overlap := 0
		for overlap < outerLen {
			if inner[innerLen-overlap-1] != outer[outerLen-overlap-1] {
				break
			}
			overlap++
		}

		if overlap > 0 {
			return append(inner[:innerLen-overlap], outer...)
		}
	}

	return append(inner, outer...)
}

// reduceStackTraces incrementally merges multiple stack traces
// and returns a merged stack trace
func reduceStackTraces(stackTraces []StackTrace) (merged StackTrace) {
	for i := len(stackTraces) - 1; i >= 0; i-- {
		merged = mergeStackTraces(merged, stackTraces[i])
	}
	return
}
