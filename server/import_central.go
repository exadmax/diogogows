// Importação do que o Diogows guardava dentro da Central Diogo (antes de virar um projeto separado): usuários, sessões, agendas,
// mensagens, contas de e-mail, perfis e planta da casa (tabelas desk_* do centraldiogo.db), os blocos "desk", "games" e "videos" do
// hub.config.json, a pasta dos arquivos dos usuários (desk/) e o games.state.json. Uso, com o serviço do Diogows parado:
//
//	diogows -config /opt/diogows/diogows.config.json -import-central /opt/centraldiogo
//
// O argumento é a pasta de instalação da Central (ou, só para o banco, o caminho de um centraldiogo.db). Nada da Central é alterado
// ou apagado: os arquivos são copiados, e a importação só roda em um Diogows ainda sem usuários.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// importFromCentral importa de uma pasta de instalação da Central (ou só o banco, se from for um arquivo).
func importFromCentral(st *store, cfgPath, from string) error {
	fi, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("não encontrei %s: %w", from, err)
	}
	if !fi.IsDir() {
		return importDB(st, from)
	}
	old := map[string]json.RawMessage{}
	if b, err := os.ReadFile(filepath.Join(from, "hub.config.json")); err == nil {
		if err := json.Unmarshal(b, &old); err != nil {
			return fmt.Errorf("hub.config.json da Central inválido: %w", err)
		}
	}
	dbPath := "centraldiogo.db"
	if raw, ok := old["dbPath"]; ok {
		var p string
		if json.Unmarshal(raw, &p) == nil && p != "" {
			dbPath = p
		}
	}
	if err := importDB(st, underDir(from, dbPath)); err != nil {
		return err
	}
	if err := mergeCentralConfig(cfgPath, old); err != nil {
		return fmt.Errorf("configuração: %w", err)
	}
	cfg := loadConfig(cfgPath)
	oldDesk := "desk"
	if raw, ok := old["desk"]; ok {
		var d struct{ Dir string }
		if json.Unmarshal(raw, &d) == nil && d.Dir != "" {
			oldDesk = d.Dir
		}
	}
	newDesk := cfg.Desk.Dir
	if newDesk == "" {
		newDesk = "desk"
	}
	newDesk = underDir(filepath.Dir(cfgPath), newDesk)
	if src := underDir(from, oldDesk); dirExists(src) && filepath.Clean(src) != filepath.Clean(newDesk) {
		if err := copyTree(src, newDesk); err != nil {
			return fmt.Errorf("copiando %s: %w", src, err)
		}
		log.Printf("importação: arquivos dos usuários copiados para %s", newDesk)
	}
	if b, err := os.ReadFile(filepath.Join(from, "games.state.json")); err == nil && cfg.Games.StateFile == "" {
		if err := os.WriteFile(filepath.Join(filepath.Dir(cfgPath), "games.state.json"), b, 0o600); err == nil {
			log.Printf("importação: games.state.json copiado")
		}
	}
	log.Printf("importação concluída. A Central não foi alterada; depois de conferir, você pode apagar %s dela.", oldDesk)
	return nil
}

// importDB copia as tabelas desk_* de um centraldiogo.db para o banco do Diogows (que precisa estar sem usuários).
func importDB(st *store, oldPath string) error {
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("banco antigo não encontrado: %w", err)
	}
	ctx := context.Background()
	if n, err := st.deskCount(ctx, ""); err != nil {
		return err
	} else if n > 0 {
		return fmt.Errorf("o banco do Diogows já tem %d usuário(s): a importação só roda em um banco novo", n)
	}
	// ATTACH vale por conexão e o banco de escrita tem uma só, então o Conn fixa a mesma durante a importação.
	conn, err := st.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ? AS antigo`, "file:"+strings.ReplaceAll(oldPath, "\\", "/")+"?mode=ro"); err != nil {
		return fmt.Errorf("abrindo o banco antigo: %w", err)
	}
	defer conn.ExecContext(ctx, `DETACH DATABASE antigo`)

	rows, err := conn.QueryContext(ctx, `SELECT name FROM antigo.sqlite_master WHERE type = 'table' AND name LIKE 'desk\_%' ESCAPE '\' ORDER BY name`)
	if err != nil {
		return err
	}
	var tables []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, n)
	}
	rows.Close()
	if len(tables) == 0 {
		return fmt.Errorf("o banco antigo não tem tabelas do Diogows (desk_*)")
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, t := range tables {
		cols, err := commonColumns(ctx, tx, t)
		if err != nil {
			return err
		}
		if len(cols) == 0 {
			log.Printf("importação: tabela %s não existe no Diogows novo, ignorada", t)
			continue
		}
		res, err := tx.ExecContext(ctx, fmt.Sprintf(`INSERT INTO main.%[1]s (%[2]s) SELECT %[2]s FROM antigo.%[1]s`, t, strings.Join(cols, ", ")))
		if err != nil {
			return fmt.Errorf("tabela %s: %w", t, err)
		}
		n, _ := res.RowsAffected()
		log.Printf("importação: %s: %d linha(s)", t, n)
	}
	return tx.Commit()
}

// commonColumns: as colunas que existem nas duas pontas, na ordem do banco novo (o antigo pode estar numa migração diferente).
func commonColumns(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	cols := func(schema string) (map[string]bool, []string, error) {
		rs, err := tx.QueryContext(ctx, fmt.Sprintf(`SELECT name FROM pragma_table_info('%s', '%s')`, table, schema))
		if err != nil {
			return nil, nil, err
		}
		defer rs.Close()
		set, order := map[string]bool{}, []string{}
		for rs.Next() {
			var n string
			if err := rs.Scan(&n); err != nil {
				return nil, nil, err
			}
			set[n] = true
			order = append(order, n)
		}
		return set, order, rs.Err()
	}
	_, novo, err := cols("main")
	if err != nil {
		return nil, err
	}
	antigo, _, err := cols("antigo")
	if err != nil {
		return nil, err
	}
	var out []string
	for _, c := range novo {
		if antigo[c] {
			out = append(out, c)
		}
	}
	return out, nil
}

// mergeCentralConfig traz "desk", "games" e "videos" do hub.config.json da Central para o diogows.config.json, sem sobrescrever
// o que o Diogows já tenha, e aponta central.url para a Central (a porta em que ela escutava). O resto da configuração da Central
// (câmeras, agendas, rotas) fica lá: nada disso passa para o Diogows.
func mergeCentralConfig(cfgPath string, old map[string]json.RawMessage) error {
	cur := map[string]json.RawMessage{}
	if b, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(b, &cur); err != nil {
			return fmt.Errorf("%s inválido: %w", cfgPath, err)
		}
	}
	for _, k := range []string{"desk", "games", "videos"} {
		if raw, ok := old[k]; ok && cur[k] == nil {
			cur[k] = raw
		}
	}
	if raw, ok := cur["desk"]; ok { // protectSettings protegia a Config da Central: acabou com a separação
		var d map[string]json.RawMessage
		if json.Unmarshal(raw, &d) == nil {
			delete(d, "protectSettings")
			cur["desk"], _ = json.Marshal(d)
		}
	}
	if cur["central"] == nil {
		listen := ":80"
		if raw, ok := old["listen"]; ok {
			json.Unmarshal(raw, &listen)
		}
		port := listen[strings.LastIndex(listen, ":")+1:]
		cur["central"], _ = json.Marshal(CentralConfig{URL: "http://127.0.0.1:" + port})
	}
	b, err := json.MarshalIndent(cur, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, append(b, '\n'), 0o600)
}

func underDir(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// copyTree copia a árvore src para dst (que não pode ter conteúdo): é o jeito seguro de levar os arquivos dos usuários sem apagar nada da origem.
func copyTree(src, dst string) error {
	if ents, err := os.ReadDir(dst); err == nil && len(ents) > 0 {
		return fmt.Errorf("%s já tem arquivos: não vou misturar", dst)
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		out := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(out, 0o700)
		}
		if !d.Type().IsRegular() { // atalhos simbólicos não são seguidos
			return nil
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, in); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	})
}
