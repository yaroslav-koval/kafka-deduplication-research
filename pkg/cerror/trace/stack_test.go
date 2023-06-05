package trace_test

import (
	"errors"
	"fmt"
	"kafka-polygon/pkg/cerror/trace"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var initpc = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() trace.Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() trace.Frame {
	return caller()
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		trace.Frame
		format string
		want   string
	}{
		{
			initpc,
			"%s",
			"stack_test.go",
		},
		{
			initpc,
			"%+s",
			"kafka-polygon/pkg/cerror/trace_test.init\n" +
				"\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go",
		},
		{
			0,
			"%s",
			trace.Unknown,
		},
		{
			0,
			"%+s",
			trace.Unknown,
		},
		{
			initpc,
			"%d",
			"14",
		},
		{
			0,
			"%d",
			"0",
		},
		{
			initpc,
			"%n",
			"init",
		},
		{
			func() trace.Frame {
				var x X
				return x.ptr()
			}(),
			"%n",
			`\(\*X\).ptr`,
		},
		{
			func() trace.Frame {
				var x X
				return x.val()
			}(),
			"%n",
			"X.val",
		},
		{
			0,
			"%n",
			"",
		},
		{
			initpc,
			"%v",
			"stack_test.go:+",
		},
		{
			initpc,
			"%+v",
			"kafka-polygon/pkg/cerror/trace_test.init\n" +
				"\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go:+",
		},
		{
			0,
			"%v",
			fmt.Sprintf("%s:0", trace.Unknown),
		},
	}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
	}
}

func TestFuncname(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"kafka-polygon/pkg/cerror/trace.funcname", "funcname"},
		{"funcname", "funcname"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := trace.FuncName(tt.name)
		want := tt.want

		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}

func TestStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want []string
	}{
		{
			trace.WithStack(errors.New("ooh")), []string{
				"kafka-polygon/pkg/cerror/trace_test.TestStackTrace\n" +
					"\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go:+",
			},
		},
		{
			trace.WithStack(errors.Join(errors.New("ooh"), errors.New("ahh"))), []string{
				"kafka-polygon/pkg/cerror/trace_test.TestStackTrace\n" +
					"\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go:+",
			},
		},
		{
			func() error { return trace.WithStack(errors.New("ooh")) }(), []string{
				`kafka-polygon/pkg/cerror/trace_test.TestStackTrace.func1` +
					"\n\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go:+", // this is the stack of New
				"kafka-polygon/pkg/cerror/trace_test.TestStackTrace\n" +
					"\t.+/kafka-polygon/pkg/cerror/trace/stack_test.go:+", // this is the stack of New's caller
			},
		},
	}
	for i, tt := range tests {
		x, ok := tt.err.(interface {
			StackTrace() trace.StackTrace
		})

		if !ok {
			t.Errorf("expected %#v to implement StackTrace() StackTrace", tt.err)
			continue
		}

		st := x.StackTrace()

		for j, want := range tt.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}

// a version of runtime.Caller that returns a Frame, not uintptr.
func caller() trace.Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])

	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()

	return trace.Frame(frame.PC)
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		trace.StackTrace
		format string
		want   string
	}{
		{
			nil,
			"%s",
			`\[\]`,
		},
		{
			nil,
			"%v",
			`\[\]`,
		},
		{
			nil,
			"%+v",
			"",
		},
		{
			nil,
			"%#v",
			`\[\]trace.Frame\(nil\)`,
		},
	}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

func testFormatRegexp(t *testing.T, n int, arg interface{}, format, want string) {
	t.Helper()

	got := fmt.Sprintf(format, arg)
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])

		if err != nil {
			t.Fatal(err)
		}

		if !match {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, got, want)
		}
	}
}
