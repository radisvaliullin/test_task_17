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

func Test_Server_Stop(t *testing.T) {

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

	// stop server
	s.Stop()
	s.Wait()

	// stop client
	// tc.Stop()
	tc.Wait()
	select {
	case err := <-tc.Error():
		t.Logf("test client running err: %v", err)
	default:
		t.Fatalf("test client stopped")
	}
}

func Test_Server_MultiClient(t *testing.T) {

	// new server init
	s := New(Config{Addr: testSrvAddr, LoginDeadline: time.Millisecond * 50, MsgDeadline: time.Millisecond * 50}, testOutLog)
	// start server
	err := s.Start()
	if err != nil {
		t.Fatalf("server start err: %v", err)
	}

	//
	clns := make([]*TestClient, 10)
	imeis := genIMEIs(10)

	// test clients
	for i, imei := range imeis {
		tc := NewTestClient(TestClientConfig{SrvAddr: testSrvAddr, IMEI: imei, PeriodDuration: time.Millisecond * 25})
		tc.Start()
		clns[i] = tc
	}

	// wait while client sent few messages
	time.Sleep(time.Millisecond * 50)

	// stop clients
	for _, tc := range clns {
		tc.Stop()
		tc.Wait()
		select {
		case err := <-tc.Error():
			t.Fatalf("test client running err: %v", err)
		default:
			t.Logf("test client stopped")
		}
	}

	// stop server
	s.Stop()
	s.Wait()
}

func BenchmarkServer(b *testing.B) {

	// new server init
	s := New(Config{Addr: testSrvAddr, LoginDeadline: time.Millisecond * 50, MsgDeadline: time.Millisecond * 50}, testOutLog)
	// start server
	err := s.Start()
	if err != nil {
		b.Fatalf("server start err: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
			b.Fatalf("test client running err: %v", err)
		default:
			b.Logf("test client stopped")
		}
	}

	// stop server
	s.Stop()
	s.Wait()
}

func BenchmarkServer_MultiClient(b *testing.B) {

	// new server init
	s := New(Config{Addr: testSrvAddr, LoginDeadline: time.Millisecond * 50, MsgDeadline: time.Millisecond * 50}, testOutLog)
	// start server
	err := s.Start()
	if err != nil {
		b.Fatalf("server start err: %v", err)
	}

	clns := make([]*TestClient, 100)
	imeis := genIMEIs(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		for i, imei := range imeis {
			// test client
			tc := NewTestClient(TestClientConfig{SrvAddr: testSrvAddr, IMEI: imei, PeriodDuration: time.Millisecond * 25})
			tc.Start()
			clns[i] = tc
		}

		// wait while client sent few messages
		time.Sleep(time.Millisecond * 50)

		for _, tc := range clns {
			// stop client
			tc.Stop()
			tc.Wait()
			select {
			case err := <-tc.Error():
				b.Fatalf("test client running err: %v", err)
			default:
				b.Logf("test client stopped")
			}
		}
	}

	// stop server
	s.Stop()
	s.Wait()
}

func Test_genIMEIs(t *testing.T) {

	for _, imei := range genIMEIs(11) {
		t.Logf("%v", imei)
	}
}

// get N IMEI
func genIMEIs(n int) [][15]byte {

	// default N
	if n < 1 || n > 1000 {
		n = 10
	}
	//
	baseIMEI := [15]byte{4, 9, 0, 1, 5, 4, 2, 0, 0, 0, 0, 0, 0, 0, 0}
	IMEIs := make([][15]byte, n)

	for i := 0; i < n; i++ {
		baseIMEI[10] = byte((i % 10000) / 1000)
		baseIMEI[11] = byte((i % 1000) / 100)
		baseIMEI[12] = byte((i % 100) / 10)
		baseIMEI[13] = byte((i % 10))

		// find check
		var check byte
		for j, b := range baseIMEI[:imeiLength-1] {
			// set value of new gen imei
			IMEIs[i][j] = b
			// each byte should be decimal (0-9)
			if b > 9 {
				panic("imei should be decimal number (0-9)")
			}
			if j%2 > 0 {
				b = b * 2
				if b > 9 {
					b = b/10 + b%10
				}
			}
			check += b
		}
		check = check % 10
		if check != 0 {
			check = 10 - check
		}
		IMEIs[i][14] = check
	}
	return IMEIs
}
