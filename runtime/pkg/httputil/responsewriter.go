package httputil

// This file was derived from: https://github.com/go-chi/chi/blob/master/middleware/wrap_writer.go

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// ResponseWriter wraps an http.ResponseWriter with response metadata tracking.
type ResponseWriter interface {
	http.ResponseWriter
	// Status returns the HTTP status of the request, or 0 if one has not
	// yet been sent.
	Status() int
	// BytesWritten returns the total number of bytes sent to the client.
	BytesWritten() int
}

// WrapResponseWriter wraps an http.ResponseWriter with response metadata tracking.
func WrapResponseWriter(w http.ResponseWriter, protoMajor int) ResponseWriter {
	_, fl := w.(http.Flusher)

	bw := baseWriter{ResponseWriter: w}

	if protoMajor == 2 {
		_, ps := w.(http.Pusher)
		if fl && ps {
			return &http2FancyWriter{bw}
		}
	} else {
		_, hj := w.(http.Hijacker)
		_, rf := w.(io.ReaderFrom)
		if fl && hj && rf {
			return &httpFancyWriter{bw}
		}
		if fl && hj {
			return &flushHijackWriter{bw}
		}
		if hj {
			return &hijackWriter{bw}
		}
	}

	if fl {
		return &flushWriter{bw}
	}

	return &bw
}

// baseWriter is the base wrapper around http.ResponseWriter that tracks metadata.
type baseWriter struct {
	http.ResponseWriter
	code        int
	bytes       int
	wroteHeader bool
}

func (b *baseWriter) WriteHeader(code int) {
	if code >= 100 && code <= 199 && code != http.StatusSwitchingProtocols {
		b.ResponseWriter.WriteHeader(code)
	} else if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}

func (b *baseWriter) Write(buf []byte) (n int, err error) {
	b.maybeWriteHeader()
	n, err = b.ResponseWriter.Write(buf)
	b.bytes += n
	return n, err
}

func (b *baseWriter) maybeWriteHeader() {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
}

func (b *baseWriter) Status() int {
	return b.code
}

func (b *baseWriter) BytesWritten() int {
	return b.bytes
}

// flushWriter is a HTTP writer that additionally satisfies http.Flusher.
type flushWriter struct {
	baseWriter
}

var _ http.Flusher = &flushWriter{}

func (f *flushWriter) Flush() {
	f.wroteHeader = true
	fl := f.baseWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

// hijackWriter is a HTTP writer that additionally satisfies http.Hijacker.
type hijackWriter struct {
	baseWriter
}

var _ http.Hijacker = &hijackWriter{}

func (f *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.baseWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

// flushHijackWriter is a HTTP writer that additionally satisfies http.Flusher and http.Hijacker.
type flushHijackWriter struct {
	baseWriter
}

var _ http.Flusher = &flushHijackWriter{}

var _ http.Hijacker = &flushHijackWriter{}

func (f *flushHijackWriter) Flush() {
	f.wroteHeader = true
	fl := f.baseWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (f *flushHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.baseWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

// httpFancyWriter is a HTTP writer that additionally satisfies http.Flusher, http.Hijacker, and io.ReaderFrom.
type httpFancyWriter struct {
	baseWriter
}

var _ http.Flusher = &httpFancyWriter{}

var _ http.Hijacker = &httpFancyWriter{}

var _ io.ReaderFrom = &httpFancyWriter{}

func (f *httpFancyWriter) Flush() {
	f.wroteHeader = true
	fl := f.baseWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (f *httpFancyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.baseWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (f *httpFancyWriter) ReadFrom(r io.Reader) (int64, error) {
	rf := f.baseWriter.ResponseWriter.(io.ReaderFrom)
	f.baseWriter.maybeWriteHeader()
	n, err := rf.ReadFrom(r)
	f.baseWriter.bytes += int(n)
	return n, err
}

// http2FancyWriter is a HTTP writer that additionally satisfies http.Flusher and io.ReaderFrom.
type http2FancyWriter struct {
	baseWriter
}

var _ http.Flusher = &http2FancyWriter{}

var _ http.Pusher = &http2FancyWriter{}

func (f *http2FancyWriter) Flush() {
	f.wroteHeader = true
	fl := f.baseWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (f *http2FancyWriter) Push(target string, opts *http.PushOptions) error {
	return f.baseWriter.ResponseWriter.(http.Pusher).Push(target, opts)
}
