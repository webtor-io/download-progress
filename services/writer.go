package services

import (
	"bufio"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

type Writer struct {
	http.ResponseWriter
	statusCode int
	sp         *StatPool
	id         string
}

func NewWriter(id string, sp *StatPool, w http.ResponseWriter) *Writer {
	return &Writer{
		statusCode:     http.StatusOK,
		ResponseWriter: w,
		sp:             sp,
		id:             id,
	}
}

func (w *Writer) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *Writer) Write(p []byte) (int, error) {
	s := w.sp.Get(w.id)
	s.Inc(int64(len(p)))
	s.status = Active
	return w.ResponseWriter.Write(p)
}

func (w *Writer) Error(err error) {
	s := w.sp.Get(w.id)
	if err != nil {
		s.status = Failed
	}
	s.status = Done
}

func (w *Writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("type assertion failed http.ResponseWriter not a http.Hijacker")
	}
	return h.Hijack()
}

func (w *Writer) Flush() {
	f, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}

// Check interface implementations.
var (
	_ http.ResponseWriter = &Writer{}
	_ http.Hijacker       = &Writer{}
	_ http.Flusher        = &Writer{}
)
