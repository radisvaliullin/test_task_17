package server

import "sync"

type readingMessage struct {
	Temp    float64
	Alt     float64
	Lat     float64
	Lon     float64
	BattLev float64
}

func (m *readingMessage) isValid() bool {
	return isRange(m.Temp, -300, 300, true) &&
		isRange(m.Alt, -20_000, 20_000, true) &&
		isRange(m.Lat, -90, 90, true) &&
		isRange(m.Lon, -180, 180, true) &&
		isRange(m.BattLev, 0, 100, false)
}

func isRange(v, min, max float64, minInclusive bool) bool {
	if v > max {
		return false
	}
	if minInclusive && v < min {
		return false
	} else if !minInclusive && v <= min {
		return false
	}
	return true
}

//
type devReq chan devReqResp
type devReqResp chan deviceReadingStatus

//
type devStorage struct {
	// map[imei]value
	mux     sync.Mutex
	storage map[string]devReq
}

func newDevStorage() *devStorage {
	ds := &devStorage{
		storage: make(map[string]devReq),
	}
	return ds
}

// setIfNot if IMEI does not exist in stor set IMEI to storage and return True or return False
func (s *devStorage) setIfNot(imei string, dr devReq) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.storage[imei]; ok {
		return false
	}
	s.storage[imei] = dr
	return true
}

func (s *devStorage) delete(imei string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.storage, imei)
}

func (s *devStorage) ok(imei string) (devReq, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	dr, ok := s.storage[imei]
	return dr, ok
}

type deviceStatus struct {
	IMEI   string `json:"imei"`
	Status string `json:"status"`
}

type deviceReadingStatus struct {
	deviceStatus
	Reading readingMessage `json:"reading,omitempty"`
	Time    int64          `json:"time,omitempty"`
}
