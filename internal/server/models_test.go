package server

import "testing"

func Test_devStorage(t *testing.T) {

	ds := newDevStorage()
	if ok := ds.setIfNot("imei"); ok {
		t.Logf("set imei to storage")
	} else {
		t.Fatalf("set imei to storage fail")
	}
	if ok := ds.setIfNot("imei"); ok {
		t.Fatalf("set imei to storage fail, exist")
	} else {
		t.Logf("set imei to storage, exist")
	}
}
