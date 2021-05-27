package services

import (
	"sync/atomic"
	"time"

	"github.com/paulbellamy/ratecounter"
)

const statInterval = 60

type Status int

const (
	Pending Status = iota
	Active
	Done
	Failed
)

type Stat struct {
	bytesWritten int64
	rate         int64
	status       Status
	cnt          *ratecounter.RateCounter
}

func NewStat() *Stat {
	return &Stat{
		cnt: ratecounter.NewRateCounter(time.Duration(statInterval) * time.Second),
	}
}

func (s *Stat) Rate() int64 {
	return s.cnt.Rate() / statInterval
}

func (s *Stat) Inc(i int64) {
	s.cnt.Incr(i)
	atomic.AddInt64(&s.bytesWritten, i)
}
