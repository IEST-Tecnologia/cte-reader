package spec

import (
	"testing"
	"time"
)

func TestCastValue(t *testing.T) {
	if got := castValue("abc", KindNumber); got != "abc" {
		t.Errorf("unparseable number = %#v, want raw string", got)
	}
	if got := castValue("00123", KindText); got != "00123" {
		t.Errorf("text = %#v, leading zeros must survive", got)
	}
	if got := castValue("2026-03-11", KindDate); got != time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC) {
		t.Errorf("date-only = %#v", got)
	}
	if got := castValue("nao-e-data", KindDate); got != "nao-e-data" {
		t.Errorf("unparseable date = %#v, want raw string", got)
	}
}
