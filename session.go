package scyna

import (
	sync "sync"
	"time"
)

type session struct {
	id       uint64
	mutex    sync.Mutex
	sequence uint64
	quit     chan struct{}
}

func NewSession(id uint64) *session {
	ret := &session{
		id:       id,
		sequence: 1,
		quit:     make(chan struct{}),
	}

	LOG = &logger{session: true, ID: id}

	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				EmitSignal(SESSION_UPDATE_CHANNEL, &UpdateSessionSignal{ID: ret.id})
			case <-ret.quit:
				ticker.Stop()
				return
			}
		}
	}()
	return ret
}

func (s *session) ID() uint64 {
	return s.id
}

func (s *session) NextSequence() uint64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sequence++
	return s.sequence
}

func (s *session) release() {
	close(s.quit)
}
