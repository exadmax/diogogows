// Uso de CPU e memória: do próprio Diogows e da máquina inteira (mostrado no aplicativo Computador). Um laço amostra a cada poucos segundos
// e as consultas só leem o último resultado, então o custo não depende de quantos painéis estão abertos.
package main

import (
	"net/http"
	"runtime"
	"sync"
	"time"
)

// sysSample é uma leitura crua do sistema; a leitura em si é por plataforma (sysinfo_linux.go, sysinfo_windows.go).
type sysSample struct {
	memTotal, memAvail uint64  // RAM da máquina, em bytes
	appRSS             uint64  // memória residente do Diogows, em bytes
	cpuIdle, cpuTotal  float64 // tempo de CPU da máquina (soma de todos os núcleos), ociosa e total, na mesma unidade
	appCPU             float64 // tempo de CPU consumido pelo Diogows, em segundos
}

type sysStats struct {
	Cores         int     `json:"cores"`         // núcleos lógicos da máquina
	CPUPercent    float64 `json:"cpuPercent"`    // uso da CPU da máquina (0-100)
	AppCPUPercent float64 `json:"appCpuPercent"` // uso do Diogows, como % da CPU total da máquina (0-100)
	MemTotal      uint64  `json:"memTotal"`      // bytes
	MemUsed       uint64  `json:"memUsed"`       // bytes em uso na máquina (total - disponível)
	MemPercent    float64 `json:"memPercent"`
	AppMem        uint64  `json:"appMem"` // bytes residentes do Diogows
	AppMemPercent float64 `json:"appMemPercent"`
}

type sysMon struct {
	mu     sync.Mutex
	prev   sysSample
	prevAt time.Time
	cur    sysStats
}

func newSysMon() *sysMon {
	m := &sysMon{}
	m.sample()
	return m
}

func (m *sysMon) loop() {
	for range time.Tick(3 * time.Second) {
		m.sample()
	}
}

func (m *sysMon) sample() {
	s, err := readSys()
	now := time.Now()
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cores := runtime.NumCPU()
	st := sysStats{Cores: cores, CPUPercent: m.cur.CPUPercent, AppCPUPercent: m.cur.AppCPUPercent, MemTotal: s.memTotal, AppMem: s.appRSS}
	if s.memTotal > 0 {
		if s.memAvail <= s.memTotal {
			st.MemUsed = s.memTotal - s.memAvail
		}
		st.MemPercent = ratioPct(float64(st.MemUsed), float64(s.memTotal))
		st.AppMemPercent = ratioPct(float64(s.appRSS), float64(s.memTotal))
	}
	if !m.prevAt.IsZero() {
		if dt := s.cpuTotal - m.prev.cpuTotal; dt > 0 {
			st.CPUPercent = ratioPct(dt-(s.cpuIdle-m.prev.cpuIdle), dt)
		}
		if wall := now.Sub(m.prevAt).Seconds(); wall > 0 {
			st.AppCPUPercent = ratioPct(s.appCPU-m.prev.appCPU, wall*float64(cores))
		}
	}
	m.prev, m.prevAt, m.cur = s, now, st
}

// ratioPct devolve a/b em porcentagem, com duas casas decimais e limitada a 0-100.
func ratioPct(a, b float64) float64 {
	if b <= 0 || a < 0 {
		return 0
	}
	p := a / b * 100
	if p > 100 {
		p = 100
	}
	return float64(int(p*100+0.5)) / 100
}

func (m *sysMon) handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	st := m.cur
	m.mu.Unlock()
	writeJSON(w, st)
}
