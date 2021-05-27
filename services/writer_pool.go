package services

import (
	"net/http"
)

type WriterPool struct {
	sp *StatPool
}

func NewWriterPool(sp *StatPool) *WriterPool {
	return &WriterPool{
		sp: sp,
	}
}

func (s *WriterPool) Get(id string, w http.ResponseWriter) *Writer {
	return NewWriter(id, s.sp, w)
}
