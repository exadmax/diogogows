# Arquitetura

```text
                        ┌────────────────────────── servidor da casa (ex.: "vovozinho") ──────────────────────────┐
 navegador              │                                                                                          │
 (PC, celular, TV)      │  ┌────────────── Diogows (:8090) ──────────────┐        ┌──── Central Diogo (:80) ────┐  │
        │               │  │ static.go  → serve www/ (gzip, cache)       │        │ painel da casa              │  │
        │  HTTP         │  │ desk*.go   → /api/desk/*  (login, arquivos, │  HTTP  │ /api/cameras  /api/contatos │  │
        ├──────────────►│  │              e-mail, chat, terminal, ...)   ├───────►│ /api/ics  /api/settings/*   │  │
        │               │  │ central.go → repassa rotas da Central       │        └─────────────────────────────┘  │
        │               │  │ games*.go  → /api/games                     │                                         │
        │               │  └───────┬──────────────────────┬──────────────┘        ┌──── Navidrome (:4533) ──────┐  │
        │               │          │ SQLite               │ arquivos              │ música (link no menu)       │  │
        │               │      diogows.db          desk/<id>/files ...            └─────────────────────────────┘  │
        │               └──────────────────────────────────────────────────────────────────────────────────────────┘
```

O Diogows é **um serviço entre outros** no servidor da casa: tem o próprio binário, porta, banco, pasta de dados e unidade systemd
(`diogows.service`), exatamente como o Navidrome ou a Central. Nada é compartilhado em disco com a Central; a única ligação é HTTP
([central.md](central.md)).

## O caminho de um pedido

1. O navegador abre `http://servidor:8090/` → `index.html` redireciona para `desktop.html`.
2. `desktop.html` carrega `js/desktop/*.js`. O `boot.js` chama `GET /api/desk/me`: sem sessão (401) mostra o login.
3. O login (`POST /api/desk/login`) confere a senha (PBKDF2), cria uma **sessão** (cookie `desk_session`, `HttpOnly`, `SameSite=Strict`) e devolve o usuário.
4. Cada janela chama a API por `DK.api(método, caminho)` (rotas `/api/desk/*`, com o cabeçalho `X-Desk: 1`) ou por `DK.hub(...)` (rotas da Central, repassadas).
5. No servidor, todo handler passa por `d.guard(admin, allowMustChange, handler)` (`desk.go`): confere cabeçalho e origem nas escritas, sessão válida,
   troca de senha pendente e perfil de administrador. Só então o handler roda, já recebendo o `*DeskUser`.
6. O handler lê/escreve no SQLite (`store`) ou na pasta de arquivos do usuário e responde JSON.

## Peças e responsabilidades

| Peça | Onde | Faz |
|---|---|---|
| **Servidor HTTP** | `main.go` | Lê a configuração, abre o banco, cria o `deskAPI`, registra as rotas e sobe o `http.Server` |
| **Configuração** | `config.go` | `diogows.config.json` (porta, front-end, banco, Central, vídeos, jogos, bloco `desk`) |
| **Banco** | `db.go` | SQLite, migrações por posição (`PRAGMA user_version`), uma conexão de escrita e um pool de leitura |
| **Autenticação e permissões** | `desk.go`, `desk_sesscache.go` | Usuários, sessões, limite de tentativas, `guard`, cache de sessão em memória |
| **Áreas funcionais** | `desk_*.go` | Cada uma registra as próprias rotas (`registerFS`, `registerMail`, `registerChat`, ...) |
| **Ponte da Central** | `central.go` | Proxy reverso com as regras de acesso |
| **Entrega do front-end** | `static.go`, `office.go` | gzip em memória, cache de um ano nos endereços `?v=VERSÃO`, cabeçalhos COOP/COEP só para `office.html` |
| **Front-end** | `www/` | A área de trabalho (veja [frontend.md](frontend.md)) |

## O que roda em segundo plano

| O quê | Quando | Onde |
|---|---|---|
| Amostra de CPU/memória | a cada 3 s | `sysinfo.go` (alimenta o aplicativo Computador) |
| Verificação de versões dos componentes | a cada 7 dias (2 min depois de subir) | `desk_updates.go` |
| Atualização do placar de jogos | a cada `games.cacheMinutes` (6 h) | `games.go` |
| Limpeza de sessões vencidas (`DELETE ... expires_at <= agora`) | ao iniciar uma sessão | `desk.go` |

Nada disso é necessário para o resto funcionar: cada tarefa falha em silêncio (e no log) sem derrubar o servidor.

## Plataformas

Linux é o alvo principal (os binários de `build.sh`: amd64, 386, arm64, arm). O **Windows** também é suportado (instalador em janela, tarefa agendada):
o que depende de `systemctl`, `sudo` ou `/proc` tem versão própria em arquivos `_windows.go`/`_other.go` (`desk_shell_*.go`, `desk_computer_*.go`, `sysinfo_*.go`, `logfile_*.go`)
ou simplesmente não existe fora do Linux (o Terminal com `sudo`).
