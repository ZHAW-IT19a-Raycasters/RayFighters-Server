package log

import "testing"

func BenchmarkLoggerPrint(b *testing.B) {
	// run the logger print n times
	clRaw := NewConsoleHandler(Debug)
	cl, _ := clRaw.Build()
	il := newInternalLogger()
	il.handlers = append(il.handlers, cl)
	for n := 0; n < b.N; n += 1 {
		il.logf(Debug, "%d", n)
	}
}
