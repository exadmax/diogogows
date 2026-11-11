# Segurança: onde cada regra vive

O modelo de ameaça e a lista completa estão no [SECURITY.md](../../SECURITY.md). Aqui, o **mapa para o código**: se vai mexer numa área, saiba qual regra ela carrega e qual teste a protege.

## Princípios

1. **Rede de casa, não internet.** Tudo foi desenhado para uma rede confiável; HTTP puro é normal e os avisos dizem isso.
2. **Negar por padrão.** Terminal, arquivos do computador, navegador, e-mail e atualizações têm interruptor e, quando perigosos (Terminal), vêm **desligados**.
3. **Cada usuário só alcança o que é dele.** Filtrar sempre por `u.ID` no servidor; nunca confiar em um id vindo do navegador sem conferir o dono.
4. **O servidor nunca executa nem interpreta o que o usuário envia.** Arquivos saem como dado inerte (`nosniff`, política de conteúdo restritiva).
5. **Segredos não saem.** A API devolve agendas e URLs mascaradas, nunca credenciais de e-mail; senhas e tokens só aparecem cifrados ou como hash.

## Regra → código

| Regra | Onde | Teste |
|---|---|---|
| Sessão em cookie `HttpOnly`, `SameSite=Strict`, `Secure` em HTTPS; só o hash do token no banco | `desk.go` (`startSession`, `deskSession*`) | `desk_test.go` |
| Escritas exigem `X-Desk: 1` + origem igual | `desk.go` (`guard`, `sameOrigin`) | `TestDeskCSRFeSessao` |
| Senha PBKDF2-SHA256, 600 mil iterações; usuário inexistente também paga o hash | `desk.go` | `desk_test.go` |
| 5 falhas por (IP, usuário) bloqueiam 5 min | `desk.go` (`deskLimiter`) | `desk_test.go` |
| Senha inicial sorteada, uma vez no log; troca obrigatória | `desk.go` (`ensureFirstInstall`, `MustChange`) | `TestDeskPrimeiraInstalacaoESenha`, `TestDeskTrocaObrigatoria` |
| Caminhos de arquivo presos à raiz do usuário | `desk_fs.go` (`localFS`) | `TestDeskHomeLinkSimbolicoNaoEscapa` |
| Terminal/arquivos do computador: ligado no servidor + permissão do admin + rede local ou HTTPS + senha própria (15 min) + senha de admin para `sudo` | `desk_shell.go` (`shellGate`) | `TestDeskComputadorPortoes`, `TestDeskSudoPedeSenhaDeAdmin` |
| Navegador só alcança a **internet** (bloqueia rede interna, o próprio servidor, metadados de nuvem); HTML sem scripts | `desk_browser.go` (`newSafeClient`), `desk_webicon.go` | `desk_browser_test.go`, `desk_webicon_test.go` |
| Agendas iCal só de endereços públicos | `desk_calendar.go` (`newCalClient`) | `desk_calendar_test.go` |
| Credenciais de e-mail cifradas (AES-GCM, chave em `desk/_mail.key`); só TLS; só internet (salvo `allowPrivate`); HTML da mensagem em iframe sem scripts e sem rede | `desk_mail*.go` | `desk_mail_test.go` |
| OAuth com PKCE e `state` de uso único | `desk_mail_oauth.go` | `desk_mail_test.go` |
| Foto de perfil só se for JPEG/PNG decodificável (≤ 400 KB), servida inerte | `desk_profile.go` | `desk_profile_test.go` |
| Planta da casa: só IPv4 da rede local (o servidor não consulta a internet nem a si mesmo) | `desk_home.go` (`homeHostOK`) | `desk_home_test.go` |
| Atualizações: só administrador; hash sha512 do pacote conferido; 3 endereços fixos, só leitura | `desk_updates.go` | `desk_updates_test.go` |
| LiteOffice: origem do código só da configuração, nunca da URL; COOP/COEP só em `office.html` | `www/js/office/office.js`, `office.go` | `office_test.go` |
| Ponte da Central: sessão, origem, administrador para `/api/settings/*`; cookie não vai | `central.go` | `TestDeskCentral*` |

## Ao acrescentar algo

- **Rota nova** → `d.guard(...)`. Pergunte: precisa ser só admin? Altera algo (então exige `X-Desk` e origem)?
- **Pede uma URL ao usuário** (e o servidor vai buscá-la) → use `newSafeClient`/`newCalClient`; nunca `http.DefaultClient`.
- **Guarda um segredo** → cifre (como o e-mail) ou guarde só o hash (como a senha); nunca devolva o valor pela API (mascare).
- **Serve um arquivo do usuário** → `nosniff` e `Content-Security-Policy` restritiva; nunca `text/html`.
- **Executa algo no servidor** → passe pelos portões de `desk_shell.go` e registre no log (`desktop-shell: ...`).

## Quem protege a Central?

Ninguém além da rede: ela não tem login. O Diogows só controla **o acesso pelo seu caminho** (veja [central.md](central.md) e o [SECURITY.md](../../SECURITY.md#a-ligação-com-a-central-diogo)).
