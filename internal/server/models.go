package server

import "sync"

type readingMessage struct {
	Temp    float64
	Alt     float64
	Lat     float64
	Lon     float64
	BattLev float64
}

//
type devStorage struct {
	// map[imei]value
	mux     sync.Mutex
	storage map[string]struct{}
}

func newDevStorage() *devStorage {
	ds := &devStorage{
		storage: make(map[string]struct{}),
	}
	return ds
}

// setIfNot if IMEI does not exist in stor set IMEI to storage and return True or return False
func (s *devStorage) setIfNot(imei string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.storage[imei]; ok {
		return false
	}
	s.storage[imei] = struct{}{}
	return true
}

func (s *devStorage) delete(imei string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.storage, imei)
}
