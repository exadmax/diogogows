# O servidor (Go)

Módulo `diogows` em `server/`. Um único pacote `main`; os arquivos são separados por **área funcional**, não por camada.

## Os arquivos

| Arquivo | Conteúdo |
|---|---|
| `main.go` | Flags (`-config`, `-xbox-login`, `-import-central`), abertura do banco, registro das rotas, `http.Server` |
| `config.go` | `Config` e `loadConfig`: padrões e leitura de `diogows.config.json` |
| `db.go` | `store`, `openStore`, `migrations`, `writeJSON`/`writeStatus`/`writeError` |
| `desk.go` | `DeskConfig`, `DeskUser`, `deskAPI`, login/logout/senha, administração de usuários, **`guard`**, limite de tentativas, `register` |
| `desk_sesscache.go` | Cache do usuário de cada sessão (1 min), invalidado quando a sessão ou o usuário mudam |
| `desk_fs.go` | Arquivos: `list`, `read`, `put`, `mkdir`, `move`, `delete` nos espaços `home` e `machine` |
| `desk_shell*.go` | Terminal e `sudo`: os portões de segurança e a execução do comando |
| `desk_computer_*.go`, `desk_misc.go` | Informações da máquina (aplicativo Computador) e lista/entrega de vídeos |
| `desk_calendar.go` | Agendas iCal por usuário, preferências (`desk_prefs`) |
| `desk_chat.go` | Mensagens entre usuários, anexos em `desk/_chat/` |
| `desk_mail*.go` | E-mail: contas, IMAP/SMTP, OAuth (Gmail, Outlook), cifra das credenciais |
| `desk_browser.go`, `desk_webicon.go` | Navegador interno, downloads, ícone de sites (cliente HTTP "seguro", só internet) |
| `desk_profile.go` | Nome, foto e assinatura |
| `desk_home.go` | Planta da casa e estado das TVs |
| `desk_vm.go` | Sistemas virtuais (pacote `.jsdos` + disco por usuário) |
| `desk_updates.go` | Atualização dos componentes de terceiros (Webamp de um clique) |
| `central.go` | Proxy da Central ([central.md](central.md)) |
| `games*.go` | Placar de jogos, sem tela (`GET /api/games`) |
| `static.go`, `office.go` | Entrega do front-end |
| `videos.go` | Lista de vídeos e legendas (`GET /api/videos/sub`) |
| `import_central.go` | `-import-central`: leva os dados da época da Central |

Cada `desk_*.go` tem, no topo, um comentário que explica a área (o que faz, regras de segurança, formato dos dados). **Comece sempre por ele.**

## Como uma rota é escrita

Todas as rotas de `/api/desk/*` seguem o mesmo molde (exemplo: `desk_profile.go`):

```go
func (d *deskAPI) registerProfile(mux *http.ServeMux) {
    mux.HandleFunc("GET /api/desk/profile", d.guard(false, true, d.profileGet))
    mux.HandleFunc("PUT /api/desk/profile", d.guard(false, true, d.profilePut))
}

func (d *deskAPI) profileGet(w http.ResponseWriter, r *http.Request, u *DeskUser) {
    p := d.st.deskProfile(r.Context(), u)
    writeStatus(w, http.StatusOK, p)
}
```

`d.guard(admin bool, allowMustChange bool, handler)`:

| Parâmetro | Significado |
|---|---|
| `admin` | `true` = só administrador (senão, 403 "somente administrador") |
| `allowMustChange` | `true` = vale mesmo com a senha provisória (só login, troca de senha, `me` e perfil usam) |

O que o `guard` faz, nesta ordem: banco disponível (senão 503) → em escritas, cabeçalho `X-Desk: 1` e origem igual (senão 403) → sessão válida (senão 401) →
senha provisória (403 `must_change`) → administrador (403). **Rota nova de `/api/desk/*` não deve ser registrada sem `guard`**, a não ser que seja deliberadamente pública
(só `login` e o retorno OAuth do e-mail são; ambos têm justificativa no código).

Erros usam `deskErr(w, status, "mensagem")` → `{"error": "mensagem"}`, que o front-end mostra como texto.

## Usuários, sessões e permissões

- **Senhas**: PBKDF2-SHA256 com 600 mil iterações (`deskPBKDF2Iter`, variável só para os testes usarem menos), sal aleatório. Usuário inexistente também paga o custo do hash.
- **Sessão**: token aleatório no cookie; no banco só o **hash** (`desk_sessions.token_hash`). Dura 12 h (`deskSessionTTL`). `Secure` quando a conexão é HTTPS.
- **Limite de tentativas**: 5 falhas por (IP, usuário) bloqueiam por 5 min (`deskLimiter`).
- **Perfis**: `is_admin` faz tudo; `can_terminal` e `can_explorer` liberam, por usuário, o Terminal e os arquivos do computador (e só se `desk.shell` estiver ligado).
  Os helpers `u.terminal()` e `u.explorer()` já incluem o administrador.
- **Primeiro acesso**: `ensureFirstInstall` cria o `admin` com senha sorteada **só quando não existe nenhum usuário**; a senha aparece uma vez no log.

## Arquivos dos usuários

`d.backend(...)` (em `desk_fs.go`) escolhe o "espaço" do pedido:

| `space` | Raiz | Quem acessa |
|---|---|---|
| `home` | `desk/<id do usuário>/files` | o próprio usuário, sempre |
| `machine` | o sistema de arquivos do servidor | quem tem permissão e passa pelo **portão do computador** (`shellGate`), com `sudo` opcional |

Os caminhos são sempre resolvidos **dentro** da raiz (`localFS`); nenhum `..` ou atalho simbólico sai dela. Arquivos servidos ao navegador
saem com `nosniff` e política de conteúdo restritiva: texto, imagem, PDF e vídeo nunca são interpretados como HTML.

## Banco de dados

Migrações em `db.go`: uma lista de strings SQL aplicada em ordem; `PRAGMA user_version` guarda quantas já rodaram. **Nunca edite uma migração antiga** — acrescente uma nova
no fim (veja a receita em [desenvolvimento.md](desenvolvimento.md#nova-migração)). Tabelas: `desk_users`, `desk_sessions`, `desk_calendars`, `desk_prefs`,
`desk_messages`, `desk_mail_accounts`, `desk_mail_apps`, `desk_profiles`, `desk_home`.

`store` tem uma conexão de **escrita** (`db`, `SetMaxOpenConns(1)`, serializa gravações) e um pool de **leitura** (`rdb`, até 3 conexões, `query_only`): no modo WAL as leituras
não esperam a gravação. Em banco em memória (testes) as duas apontam para a mesma conexão.

## Entrega do front-end (`static.go`)

- Texto (`.html`, `.js`, `.css`, `.json`, `.svg`) é comprimido uma vez e guardado em memória (teto de 8 MB).
- Com um arquivo `VERSION` instalado, as páginas `.html` ganham `?v=VERSÃO` nos scripts e estilos, e esses endereços recebem cache de **um ano**;
  cada instalação muda a versão, então o navegador baixa só o que mudou. Sem `VERSION` (`dev`) tudo é `no-cache`.
- `config/` e `vendor/` ficam fora do versionamento (a primeira é editada à mão; as bibliotecas de terceiros descobrem a própria pasta pelo endereço do script).
- `office.go` acrescenta `Cross-Origin-Opener-Policy` e `Cross-Origin-Embedder-Policy` **só** a `office.html` e seus scripts (o LibreOffice em WebAssembly exige).

## Testes

Cada área tem seu `_test.go`. Os testes do `desk` usam um servidor em memória (`newDK`), com `deskPBKDF2Iter` baixo, e cobrem permissões, sessão, arquivos, terminal, e-mail (contra
um IMAP/SMTP de teste em TLS), chat, calendário, navegador e a ponte da Central (contra uma "Central" de mentira em `httptest`). `go test ./...` roda tudo em poucos segundos.
