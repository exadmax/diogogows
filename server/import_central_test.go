package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// O banco antigo é o centraldiogo.db de antes da separação: as tabelas do Diogows vêm junto com as da Central.
func TestImportDB(t *testing.T) {
	dir := t.TempDir()
	old, err := openStore(filepath.Join(dir, "centraldiogo.db")) // mesmas tabelas desk_*
	if err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`CREATE TABLE contatos (id INTEGER PRIMARY KEY, nome TEXT); INSERT INTO contatos VALUES (1, 'só da Central');
		INSERT INTO desk_users (id, username, pass_hash, is_admin, created_at) VALUES (7, 'diogo', 'hash', 1, 100)`); err != nil {
		t.Fatal(err)
	}
	old.Close()

	novo, err := openStore(filepath.Join(dir, "diogows.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer novo.Close()
	if err := importDB(novo, filepath.Join(dir, "centraldiogo.db")); err != nil {
		t.Fatal(err)
	}
	u, err := novo.deskUserBy(t.Context(), "username", "diogo")
	if err != nil || u.ID != 7 || !u.IsAdmin || u.PassHash != "hash" {
		t.Fatalf("usuário importado: %+v %v", u, err)
	}
	// importar de novo (banco já com usuários) é recusado, e tabelas da Central não vêm junto
	if err := importDB(novo, filepath.Join(dir, "centraldiogo.db")); err == nil {
		t.Error("a segunda importação deveria ser recusada")
	}
	if _, err := novo.db.Exec(`SELECT 1 FROM contatos`); err == nil {
		t.Error("tabelas da Central não deveriam existir no banco do Diogows")
	}
}

// Importação completa a partir da pasta da Central: banco, blocos de configuração e arquivos dos usuários.
func TestImportFromCentralPasta(t *testing.T) {
	root := t.TempDir()
	central := filepath.Join(root, "central")
	os.MkdirAll(filepath.Join(central, "desk", "7"), 0o755)
	os.WriteFile(filepath.Join(central, "desk", "7", "nota.txt"), []byte("oi"), 0o644)
	os.WriteFile(filepath.Join(central, "hub.config.json"), []byte(`{"listen": ":80", "cameras": [{"url": "rtsp://segredo"}],
		"desk": {"shell": true, "protectSettings": true, "maxUploadMB": 64}, "games": {"steam": {"user": "x"}}}`), 0o600)
	old, err := openStore(filepath.Join(central, "centraldiogo.db"))
	if err != nil {
		t.Fatal(err)
	}
	old.db.Exec(`INSERT INTO desk_users (id, username, pass_hash, created_at) VALUES (3, 'ana', 'h', 1)`)
	old.Close()

	novo := filepath.Join(root, "diogows")
	os.MkdirAll(novo, 0o755)
	cfgPath := filepath.Join(novo, "diogows.config.json")
	st, err := openStore(filepath.Join(novo, "diogows.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := importFromCentral(st, cfgPath, central); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(novo, "desk", "7", "nota.txt"))
	if string(b) != "oi" {
		t.Errorf("arquivos dos usuários: %q", b)
	}
	cfg := loadConfig(cfgPath)
	if !cfg.Desk.Shell || cfg.Desk.MaxUploadMB != 64 || cfg.Games.Steam.User != "x" || cfg.Central.URL != "http://127.0.0.1:80" {
		t.Errorf("configuração importada: %+v", cfg)
	}
	raw, _ := os.ReadFile(cfgPath)
	if strings.Contains(string(raw), "rtsp://segredo") || strings.Contains(string(raw), "protectSettings") {
		t.Errorf("a configuração da Central (câmeras) não pode vir junto: %s", raw)
	}
	if _, err := os.Stat(filepath.Join(central, "desk", "7", "nota.txt")); err != nil {
		t.Error("a origem não pode ser apagada")
	}
}
