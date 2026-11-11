// Configuração do Diogows (diogows.config.json). Tudo é opcional: sem o arquivo o servidor sobe com os padrões.
package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Listen    string        `json:"listen"`    // endereço de escuta (padrão ":8090")
	StaticDir string        `json:"staticDir"` // pasta do front-end (padrão "../www", relativa ao diogows.config.json)
	DBPath    string        `json:"dbPath"`    // banco SQLite (padrão diogows.db, relativo ao diogows.config.json)
	Central   CentralConfig `json:"central"`   // ligação com a Central Diogo (câmeras, contatos, agendas da casa)
	Videos    VideosConfig  `json:"videos"`    // pasta dos vídeos (aplicativo Vídeos)
	Games     GamesConfig   `json:"games"`     // placar de jogos: Steam, Xbox, PlayStation e RetroAchievements (games.go)
	Desk      DeskConfig    `json:"desk"`      // usuários, arquivos, terminal, e-mail (desk.go)
}

// CentralConfig liga o Diogows à Central Diogo (o hub do painel da casa), que continua sendo outro programa, em outro serviço.
// O Diogows repassa, só para quem está logado, as câmeras, os contatos, as agendas e o menu da Central (central.go).
type CentralConfig struct {
	URL string `json:"url"` // endereço do hub da Central, ex.: http://127.0.0.1:8080; vazio desliga a ligação
}

// homeDir é a pasta pessoal de quem roda o Diogows (base do padrão da pasta de vídeos).
func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil && h != "" {
		return h
	}
	return "."
}

func loadConfig(path string) Config {
	c := Config{Listen: ":8090", StaticDir: "../www", DBPath: "diogows.db", Videos: VideosConfig{Dir: filepath.Join(homeDir(), "Videos")}}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Printf("%s não existe: usando os padrões (copie diogows.config.example.json para ajustar).", path)
		return c
	}
	if err != nil {
		log.Fatalf("não foi possível ler %s: %v", path, err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		log.Fatalf("%s inválido: %v", path, err)
	}
	return c
}
