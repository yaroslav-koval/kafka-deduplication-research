package trace_test

import (
	"io"
	"kafka-polygon/pkg/cerror/trace"
	"testing"
)

func TestWithStackNil(t *testing.T) {
	got := trace.WithStack(nil)
	if got != nil {
		t.Errorf("WithStack(nil): got %#v, expected nil", got)
	}
}

func TestWithStack(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{trace.WithStack(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := trace.WithStack(tt.err).Error()
		if got != tt.want {
			t.Errorf("WithStack(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}
