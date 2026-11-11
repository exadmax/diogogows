// Diogows: uma área de trabalho no navegador (janelas, menu Iniciar, vários usuários, e-mail, documentos, emuladores...).
// Servidor único em Go, com poucas dependências, feito para rodar em máquinas de 2 GB ao lado de outros serviços da casa
// (como o Navidrome e a Central Diogo). Serve o front-end (pasta www/) e a API /api/desk/*.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// withNoCache: o Diogows responde num endereço só (sem pedidos de outras origens); aqui só marca as páginas como sempre atuais.
func withNoCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") { // front-end sempre atualizado após uma instalação
			w.Header().Set("Cache-Control", "no-cache")
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	cfgPath := flag.String("config", "diogows.config.json", "caminho do arquivo de configuração")
	xboxLogin := flag.Bool("xbox-login", false, "autoriza a conta Xbox (uma vez) e sai")
	importCentral := flag.String("import-central", "", "importa os dados do Diogows da pasta de instalação da Central Diogo (ou de um centraldiogo.db) e sai")
	flag.Parse()
	cfgDir := filepath.Dir(*cfgPath)
	cfg := loadConfig(*cfgPath)
	if *xboxLogin {
		xboxLoginCLI(cfg, *cfgPath)
		return
	}

	if !filepath.IsAbs(cfg.DBPath) { // relativo à pasta do diogows.config.json
		cfg.DBPath = filepath.Join(cfgDir, cfg.DBPath)
	}
	if !filepath.IsAbs(cfg.StaticDir) {
		cfg.StaticDir = filepath.Join(cfgDir, cfg.StaticDir)
	}
	// Se o banco não abrir, o servidor segue no ar e /api/desk/* responde 503 (o log explica).
	db, err := openStore(cfg.DBPath)
	if err != nil {
		log.Printf("banco de dados %s: %v", cfg.DBPath, err)
		db = nil
	}
	if *importCentral != "" {
		if db == nil {
			log.Fatal("importação cancelada: o banco do Diogows não abriu")
		}
		if err := importFromCentral(db, *cfgPath, *importCentral); err != nil {
			log.Fatalf("importação falhou: %v", err)
		}
		return
	}

	initLogFile(cfgDir) // só o servidor escreve no arquivo de log (os modos de linha de comando mostram tudo no terminal)
	mux := http.NewServeMux()
	sys := newSysMon()
	go sys.loop()
	desk := newDesk(cfg, db, sys, cfgDir)
	desk.ensureFirstInstall(context.Background())
	desk.register(mux)
	desk.registerCentral(mux, cfg.Central.URL)
	desk.startUpdates()
	mux.HandleFunc("GET /api/videos/sub", subHandler(cfg.Videos))

	gm := newGames(cfg, *cfgPath)
	go gm.loop()
	mux.HandleFunc("/api/games", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, gm.get(r.URL.Query().Get("refresh") == "1"))
	})
	registerXboxLogin(mux, gm)

	// Versão instalada: arquivo VERSION ao lado do diogows.config.json (o install.sh soma 1 ao terceiro número a cada instalação).
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		v, _ := readVersion(cfgDir)
		writeJSON(w, map[string]interface{}{"ok": true, "time": time.Now().Format(time.RFC3339), "version": v})
	})
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		v, at := readVersion(cfgDir)
		writeJSON(w, map[string]interface{}{"version": v, "installedAt": at})
	})
	mux.Handle("/", withOfficeIsolation(newStaticHandler(cfg.StaticDir, func() string { v, _ := readVersion(cfgDir); return v })))

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           withNoCache(mux),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ver, _ := readVersion(cfgDir)
	if ver != "dev" {
		ver = "V" + ver
	}
	log.Printf("Diogows %s em %s (front-end: %s)", ver, cfg.Listen, cfg.StaticDir)
	log.Fatal(srv.ListenAndServe())
}
