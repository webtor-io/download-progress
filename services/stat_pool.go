package services

import (
	"sync"
	"time"
)

const (
	statTTL = 60
)

type StatPool struct {
	expire time.Duration
	sm     sync.Map
	timers sync.Map
}

func NewStatPool() *StatPool {
	return &StatPool{
		expire: time.Duration(statTTL) * time.Second,
	}
}

func (s *StatPool) Get(id string) *Stat {
	key := id
	v, _ := s.sm.LoadOrStore(key, NewStat())
	t, tLoaded := s.timers.LoadOrStore(key, time.NewTimer(s.expire))
	timer := t.(*time.Timer)
	if !tLoaded {
		go func() {
			<-timer.C
			s.sm.Delete(key)
			s.timers.Delete(key)
		}()
	} else {
		timer.Reset(s.expire)
	}
	return v.(*Stat)
}

func (s *StatPool) GetIfExists(id string) *Stat {
	key := id
	v, loaded := s.sm.Load(key)
	if loaded {
		return v.(*Stat)
	} else {
		return nil
	}
}
