// Diogows: Meu Computador (informações da máquina) e a biblioteca de vídeos do player.
package main

import (
	"bufio"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type deskDisk struct {
	Mount  string `json:"mount"`
	Device string `json:"device,omitempty"`
	FS     string `json:"fs,omitempty"`
	Total  uint64 `json:"total"`
	Free   uint64 `json:"free"`
}

type deskComputer struct {
	Hostname   string     `json:"hostname"`
	OS         string     `json:"os"`
	Platform   string     `json:"platform,omitempty"` // nome da distribuição (Linux)
	Arch       string     `json:"arch"`
	CPUs       int        `json:"cpus"`
	CPUModel   string     `json:"cpu_model,omitempty"`
	CPUPercent float64    `json:"cpu_percent"`
	MemTotal   uint64     `json:"mem_total"`
	MemUsed    uint64     `json:"mem_used"`
	UptimeSec  float64    `json:"uptime_sec,omitempty"`
	Load       string     `json:"load,omitempty"`
	Disks      []deskDisk `json:"disks"`
	Addresses  []string   `json:"addresses"`
	HubVersion string     `json:"hub_version"`
	HubMem     uint64     `json:"hub_mem"`
	Time       string     `json:"time"`
}

func (d *deskAPI) registerMisc(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/desk/computer", d.guard(false, false, d.handleComputer))
	mux.HandleFunc("GET /api/desk/videos", d.guard(false, false, d.handleVideos))
	mux.HandleFunc("GET /api/desk/videos/file", d.guard(false, false, d.handleVideoFile))
}

func (d *deskAPI) handleComputer(w http.ResponseWriter, r *http.Request, u *DeskUser) {
	host, _ := os.Hostname()
	c := deskComputer{Hostname: host, OS: runtime.GOOS, Arch: runtime.GOARCH, CPUs: runtime.NumCPU(), Time: time.Now().Format(time.RFC3339),
		Disks: deskDisks(), Addresses: []string{}}
	c.Platform, c.CPUModel = deskPlatform()
	c.UptimeSec = deskUptime()
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		if f := strings.Fields(string(b)); len(f) >= 3 {
			c.Load = strings.Join(f[:3], " ")
		}
	}
	if d.sys != nil {
		d.sys.mu.Lock()
		s := d.sys.cur
		d.sys.mu.Unlock()
		c.CPUPercent, c.MemTotal, c.MemUsed, c.HubMem = s.CPUPercent, s.MemTotal, s.MemUsed, s.AppMem
	}
	v, _ := readVersion(d.verDir)
	c.HubVersion = v
	if ifs, err := net.Interfaces(); err == nil {
		for _, ifc := range ifs {
			if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, _ := ifc.Addrs()
			for _, a := range addrs {
				if ipn, ok := a.(*net.IPNet); ok && ipn.IP.To4() != nil {
					c.Addresses = append(c.Addresses, ifc.Name+" "+ipn.IP.String())
				}
			}
		}
	}
	writeStatus(w, http.StatusOK, c)
}

// readLinuxFile lê um arquivo de /proc ou /etc; vazio se não existir.
func readLinuxFile(name string) string {
	b, err := os.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(b)
}

func firstLineValue(text, key string) string {
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), ":"); ok && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ---------- vídeos ----------

func (d *deskAPI) handleVideos(w http.ResponseWriter, r *http.Request, u *DeskUser) {
	if d.videos.Dir == "" {
		deskErr(w, http.StatusServiceUnavailable, "pasta de vídeos não configurada no Diogows (campo \"videos.dir\")")
		return
	}
	list, err := listVideos(d.videos.Dir)
	if err != nil {
		deskErr(w, http.StatusServiceUnavailable, "não foi possível ler a pasta de vídeos do servidor")
		return
	}
	writeStatus(w, http.StatusOK, list)
}

// handleVideoFile entrega um vídeo da biblioteca com Range. O nome vindo da URL só vale se for um item da listagem: nunca
// vira caminho de arquivo por conta própria.
func (d *deskAPI) handleVideoFile(w http.ResponseWriter, r *http.Request, u *DeskUser) {
	name := r.URL.Query().Get("name")
	list, err := listVideos(d.videos.Dir)
	if d.videos.Dir == "" || err != nil {
		deskErr(w, http.StatusServiceUnavailable, "pasta de vídeos indisponível")
		return
	}
	for _, v := range list {
		if v.Name != name {
			continue
		}
		ct, ok := deskViewTypes[strings.ToLower(path.Ext(v.Name))]
		if !ok {
			ct = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})
		be := localFS{root: filepath.Clean(d.videos.Dir)}
		_ = serveBackendFile(w, r, be, "/"+v.Name, v.Size)
		return
	}
	deskErr(w, http.StatusNotFound, "vídeo não encontrado")
}
