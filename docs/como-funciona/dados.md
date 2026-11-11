# Onde ficam os dados

Tudo que o Diogows guarda está em **uma pasta** (no Linux, `/opt/diogows`; no Windows, `%ProgramData%\Diogows` ou `%LOCALAPPDATA%\Diogows`). Fazer backup dessa pasta (com o serviço
parado, ou usando a cópia do SQLite) é fazer backup de tudo.

```text
/opt/diogows/
├── diogows                    o programa (substituído a cada instalação)
├── www/                       o front-end (refeito a cada instalação)
├── diogows.config.json        a configuração (preservada; .bak a cada instalação)
├── diogows.db  (+ -wal, -shm) o banco SQLite
├── VERSION, install.log       versão instalada e histórico das instalações
├── games.state.json           tokens renovados e cache do placar de jogos (modo 0600)
├── diogows.log                só no Windows: o log (no Linux vai para o journal do systemd)
└── desk/                      a pasta dos usuários ("desk.dir" muda o lugar)
    ├── <id do usuário>/
    │   ├── files/             "Meus arquivos": Documentos, Downloads, Jogos, Músicas, Programas...
    │   └── Recebidos/         anexos recebidos no chat
    ├── _chat/                 conteúdo dos anexos do chat (nome sorteado)
    ├── _mail.key              chave (32 bytes, 0600) que cifra as credenciais de e-mail
    ├── _updates.json          resultado da última verificação de versões
    └── _vendor/               arquivos de terceiros baixados pela tela Atualizações (ex.: Webamp)
```

## O que o instalador toca

| Item | Instalação | Atualização |
|---|---|---|
| `diogows`, `www/` | criados | **substituídos** |
| `diogows.config.json` | criado (primeira vez) | **mantido** (cópia em `.bak`) |
| `diogows.db`, `desk/`, `games.state.json` | não existem | **nunca tocados** |
| `VERSION`, `install.log` | criados | atualizados |

## Banco de dados (`diogows.db`)

| Tabela | Guarda |
|---|---|
| `desk_users` | usuários: nome de login, hash da senha, perfis (`is_admin`, `can_terminal`, `can_explorer`), `must_change` |
| `desk_sessions` | sessões abertas (só o **hash** do token) |
| `desk_calendars` | agendas iCal de cada usuário (a URL é segredo: só o dono vê, mascarada) |
| `desk_prefs` | preferências da conta (aparência, menu Iniciar, programas instalados, categorias) |
| `desk_messages` | mensagens entre usuários (texto; anexos ficam em `desk/_chat/`) |
| `desk_mail_accounts`, `desk_mail_apps` | contas de e-mail (credenciais **cifradas**) e o aplicativo OAuth de cada usuário |
| `desk_profiles` | nome, sobrenome, foto, assinatura |
| `desk_home` | planta da casa (andares, cômodos, TVs) |

**O e-mail em si não fica aqui**: as mensagens ficam no provedor e o Diogows as lê por IMAP.

## Backup e restauração

- **Banco**: com o serviço rodando, use `sqlite3 diogows.db ".backup copia.db"` (o modo WAL deixa arquivos `-wal`/`-shm` ao lado; copiar só o `.db` pode perder gravações recentes).
- **Arquivos**: copie `desk/` inteira. **Inclua `desk/_mail.key`**: sem ela as contas de e-mail precisam ser ligadas de novo.
- **Restaurar**: pare o serviço, ponha `diogows.db` e `desk/` no lugar, inicie. As migrações pendentes rodam sozinhas na subida.

## Remover um usuário

Apaga as linhas dele (sessões, agendas, mensagens, e-mail, perfil) **e a pasta `desk/<id>/`**. Não há lixeira.

## Configuração (`diogows.config.json`)

Tudo é opcional. Os blocos e campos estão no [README](../../README.md#configuração) e, para o bloco `desk`, em [desktop.md](../desktop.md#configuração-diogowsconfigjson).
