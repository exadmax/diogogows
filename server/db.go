// Banco SQLite do Diogows: usuários, sessões, agendas, mensagens, e-mail, perfil e planta da casa.
// Driver em Go puro (modernc.org/sqlite), sem CGO: o Diogows continua compilando para todas as arquiteturas Linux.
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "modernc.org/sqlite"
)

// Cada entrada roda uma vez, em ordem (PRAGMA user_version guarda quantas já foram aplicadas).
// Para mudar o esquema depois, acrescente uma migração nova; nunca edite as antigas.
var migrations = []string{
	deskMigration,         // usuários e sessões (desk.go)
	deskCalendarMigration, // agendas iCal por usuário e preferências (desk_calendar.go)
	deskChatMigration,     // mensagens entre usuários (desk_chat.go)
	deskMailMigration,     // contas de e-mail por usuário (desk_mail.go)
	deskProfileMigration,  // nome, foto e assinatura dos usuários (desk_profile.go)
	deskMailAppsMigration, // aplicativo OAuth de e-mail por usuário (desk_mail.go)
	deskHomeMigration,     // planta da casa e aparelhos inteligentes (desk_home.go)
}

type store struct {
	db   *sql.DB // única conexão, para gravar (e para o que precisa ler o que acabou de gravar numa transação)
	rdb  *sql.DB // até readConns conexões só de leitura: no modo WAL elas não esperam a gravação (nem a gravação espera por elas)
	sess sessCache
}

const readConns = 3

func openStore(path string) (*store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // uso doméstico: uma conexão basta para gravar, serializa as gravações e poupa memória
	s := &store{db: db, rdb: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if path != "" && !strings.Contains(path, ":memory:") { // banco em memória é por conexão: nele a leitura usa a mesma conexão
		rdb, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=query_only(1)")
		if err != nil {
			db.Close()
			return nil, err
		}
		rdb.SetMaxOpenConns(readConns)
		s.rdb = rdb
	}
	return s, nil
}

// Close encerra as conexões de gravação e de leitura.
func (s *store) Close() error {
	err := s.db.Close()
	if s.rdb != s.db {
		if e := s.rdb.Close(); err == nil {
			err = e
		}
	}
	return err
}

func (s *store) migrate() error {
	var v int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		return err
	}
	for i := v; i < len(migrations); i++ {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migração %d: %w", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// ---------- respostas JSON ----------

// apiError vira resposta HTTP com a mensagem em JSON ({"error": "..."}) para a tela mostrar.
type apiError struct {
	status int
	msg    string
}

func (e *apiError) Error() string { return e.msg }

func badRequest(format string, a ...interface{}) error {
	return &apiError{http.StatusBadRequest, fmt.Sprintf(format, a...)}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func writeStatus(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err error) {
	var ae *apiError
	if !errors.As(err, &ae) {
		log.Printf("erro no banco de dados: %v", err)
		ae = &apiError{http.StatusInternalServerError, "erro interno ao acessar o banco de dados"}
	}
	writeStatus(w, ae.status, map[string]string{"error": ae.msg})
}
