package server

import (
	"testing"
	"time"
)

const (
	testSrvAddr = "0.0.0.0:1338"
)

func Test_Server(t *testing.T) {

	// new server init
	s := New(Config{Addr: testSrvAddr, LoginDeadline: time.Millisecond * 50, MsgDeadline: time.Millisecond * 50}, testOutLog)
	// start server
	err := s.Start()
	if err != nil {
		t.Fatalf("server start err: %v", err)
	}

	// test client
	tc := NewTestClient(TestClientConfig{SrvAddr: testSrvAddr, IMEI: testIMEIArr, PeriodDuration: time.Millisecond * 25})
	tc.Start()

	// wait while client sent few messages
	time.Sleep(time.Millisecond * 50)

	// stop client
	tc.Stop()
	tc.Wait()
	select {
	case err := <-tc.Error():
		t.Fatalf("test client running err: %v", err)
	default:
		t.Logf("test client stopped")
	}

	// stop server
	s.Stop()
	s.Wait()

}
