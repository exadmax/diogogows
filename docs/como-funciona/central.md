# A ligação com a Central Diogo

A [Central Diogo](https://github.com/exadmax/DVRCentral) é o painel da casa (agenda, clima, trânsito, câmeras, estoque, livros). O Diogows e ela **moravam no mesmo programa**;
hoje são dois, e esta página explica o que os une.

## O que o Diogows pega da Central

| Recurso | Rota da Central | Onde aparece no Diogows |
|---|---|---|
| Lista e vídeo das câmeras | `GET /api/cameras`, `/api/cameras/{id}/stream`, `/snapshot.jpg` | Aplicativo **Câmeras** |
| Configurar as câmeras | `GET/PUT /api/settings/cameras` | Câmeras → **Gerenciar** (só administrador) |
| Contatos / profissionais | `GET/POST/PUT/DELETE /api/contatos` | Aplicativo **Contatos** |
| Agendas da casa | `GET /api/ics` | **Calendário** (fontes "Casa · ...") |
| Menu e opções do painel | `GET /api/settings/menu`, `/api/settings/panel` | **Menu Iniciar** (atalhos do painel); esconde Livros se a Central os desligou |
| Páginas da Central | `config.html`, `camaras.html`, `index.html` | Links e a janela *Configurações* (um iframe) — abrem direto no endereço da Central |

Só isso. O resto da API da Central (estoque, gravações, trânsito, energia) **não** é repassado.

## Como funciona: um proxy com regras

`server/central.go` registra, quando existe `central.url`, as rotas acima como um **proxy reverso** (`httputil.ReverseProxy`) para a Central. O navegador continua
falando só com o endereço do Diogows; por isso o front-end usa caminhos relativos (`DK.hub('GET', '/api/cameras')`) e não precisa de CORS.

```text
navegador ──/api/cameras──► Diogows (central.go) ──/api/cameras──► Central
              cookie de sessão  │  confere sessão, origem e perfil      (sem cookie, sem Origin)
```

| Regra | Onde |
|---|---|
| **Leitura**: exige sessão válida e senha não provisória (401 caso contrário) | `centralProxy.ServeHTTP` |
| **Escrita** (qualquer método que altera): exige também **origem igual** à do Diogows (403) | idem |
| **`/api/settings/*` com escrita**: exige **administrador** (403) | idem |
| O cookie, `Origin`, `Referer` e `X-Desk` **não** vão para a Central | `Rewrite` |
| Nas escritas o proxy acrescenta `X-CD-Action: confirm`, que a Central exige nas gravações de configuração (a defesa dela contra formulários de outros sites) | `ServeHTTP` |
| Fluxos de câmera (`.../stream`) saem sem o limite de escrita do servidor e sem buffer (`FlushInterval: -1`) | `ServeHTTP`, `newCentralProxy` |
| Central fora do ar → 502 com a mensagem "a Central Diogo não respondeu" | `ErrorHandler` |
| Sem `central.url`, **nenhuma** dessas rotas existe (404) e os aplicativos mostram "indisponível" | `registerCentral` |

O acréscimo de `X-CD-Action` é seguro porque ele só chega depois das três checagens acima: o objetivo desse cabeçalho na Central é impedir que **outra página** dispare a
escrita; aqui quem escreve já provou sessão, origem e (para configuração) ser administrador.

## Configuração

No `diogows.config.json`:

```json
"central": {"url": "http://127.0.0.1:80"}
```

- No servidor da casa a Central escuta na **porta 80** (instalador dela) e o Diogows na **8090**; os dois falam por `127.0.0.1`, sem sair da máquina.
- No seu computador de desenvolvimento a Central costuma estar em `:8080` (`diogows.config.example.json` já vem assim).

No navegador, os **links** para páginas da Central usam `www/config/settings.js` (`central: { port: 80 }` por padrão, ou `url` completa): é separado porque o link é aberto pelo
navegador de quem usa, que pode alcançar a Central por outro nome de máquina.

## Do lado da Central

- `/desktop.html` na Central **redireciona** para o Diogows (porta `diogowsPort`, padrão 8090), para favoritos antigos e atalhos já salvos no menu da Config.
- O menu padrão da Central traz o atalho "Diogows" (porta 8090, caminho `/`).
- A Central **não sabe** que o Diogows existe além disso: nenhuma rota dela depende dele.

## Se a Central mudar

O contrato é o das rotas da tabela acima. Mudou o formato de `/api/cameras`, `/api/contatos`, `/api/ics` ou `/api/settings/*`? Ajuste `www/js/desktop/` (cameras.js, contacts.js,
calendar.js, boot.js) e rode os testes de `desk_test.go` (`TestDeskCentral*`), que usam uma Central de mentira para conferir as regras de acesso.
