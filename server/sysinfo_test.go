package main

import (
	"testing"
	"time"
)

func TestSysMon(t *testing.T) {
	m := newSysMon()
	x := 0
	for end := time.Now().Add(300 * time.Millisecond); time.Now().Before(end); x++ { // gasta CPU para o delta aparecer
	}
	m.sample()
	st := m.cur
	if st.Cores < 1 || st.MemTotal == 0 || st.AppMem == 0 || st.MemUsed == 0 || st.MemUsed > st.MemTotal {
		t.Fatalf("leitura inválida: %+v", st)
	}
	if st.AppCPUPercent <= 0 || st.AppCPUPercent > 100 || st.CPUPercent <= 0 {
		t.Fatalf("CPU inválida: %+v", st)
	}
	t.Logf("%+v", st)
}

func TestRatioPct(t *testing.T) {
	if ratioPct(1, 3) != 33.33 || ratioPct(5, 0) != 0 || ratioPct(9, 3) != 100 || ratioPct(-1, 3) != 0 {
		t.Fatal("pct")
	}
}
