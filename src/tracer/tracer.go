package tracer

import (
	"fmt"
	"io"
	"time"
)

type Tracer interface {
	Trace(...interface{})
	Tracef(format string, args ...interface{})
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprint(a...)
	t.out.Write([]byte(fmt.Sprintf("[%s] %s\n", timestamp, message)))
}

func (t *tracer) Tracef(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	t.out.Write([]byte(fmt.Sprintf("[%s] %s\n", timestamp, message)))
}

func New(w io.Writer) Tracer {
	return &tracer{
		out: w,
	}
}