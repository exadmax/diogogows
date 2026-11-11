# Desenvolvimento

## Requisitos

- **Go** (a versão está em `server/go.mod`). Só isso para compilar e testar.
- **Node.js** apenas para `node scripts/check-i18n.js` (opcional).
- Nenhum build do front-end: edite `www/` e recarregue a página.

## Rodar no seu computador

```bash
cd server
cp diogows.config.example.json diogows.config.json     # arquivo ignorado pelo git
go run .
```

Abra <http://localhost:8090>. Na primeira vez o **log** mostra a senha inicial do `admin` (no Windows, o log vai para `diogows.log`, ao lado da configuração).
O banco (`diogows.db`) e a pasta `desk/` nascem em `server/`, ignorados pelo git.

Para ver as câmeras, contatos e agendas, suba também a [Central Diogo](https://github.com/exadmax/DVRCentral) (`go run .` na pasta dela, porta 8080, que é o que
`central.url` do exemplo já espera). Sem ela, tudo funciona, e esses itens mostram "indisponível".

**Instância de demonstração** (dados separados, sem expor nada pessoal): `go run . -config demo/diogows.config.json`, na porta 8091 ([server/demo/README.md](../../server/demo/README.md)).

### Front-end em dev

O servidor entrega `www/` direto do disco; sem arquivo `VERSION`, tudo sai com `no-cache` e basta recarregar. Se a Central do seu PC estiver em outra porta que a
padrão do front-end (80), crie `www/config/settings.js`:

```js
window.CD_SETTINGS_LOCAL = { central: { port: 8080 } };
```

## Testes

```bash
cd server
go vet ./...
go test ./...            # alguns segundos
node ../scripts/check-i18n.js
```

Os testes do `desk` sobem um servidor em memória com `newDK(t, DeskConfig{...})` e fazem pedidos como o navegador faria (`e.req(método, caminho, corpo, cookie)`).
`e.addUser(...)` cria usuários e `e.login(...)` devolve o cookie. Use esses helpers em testes novos: eles já baixam o custo da senha.

Para a ponte da Central, `centralEnv(t)` sobe uma Central de mentira e liga o Diogows a ela.

## Receitas

### Novo aplicativo (só front-end)

1. Crie `www/js/desktop/meuapp.js`:
   ```js
   (function (DK) {
     const el = DK.el;
     DK.meuApp = function () {
       DK.openWindow('meuapp', 'Meu aplicativo', body => {
         body.append(el('div', { textContent: 'Olá, ' + DK.me.username }));
       }, 520, 360);
     };
     DK.register({ id: 'meuapp', icon: '🧩', label: 'Meu aplicativo', desktop: true, open: DK.meuApp });
   })(window.DK = window.DK || {});
   ```
2. Inclua `<script src="js/desktop/meuapp.js"></script>` no `desktop.html` (ou, se for pesado, faça sob demanda: [frontend.md](frontend.md#carregamento-sob-demanda-lazyjs)).
3. Para o app aparecer no menu Iniciar, na categoria certa, acrescente o `id` em `MENU_ORDER` e na lista de `CATS` em `boot.js`; sem isso ele cai em "Outros".
4. Recarregue. Use `DK.api('GET', '/rota')` para falar com o servidor.

### Nova rota na API

1. Escolha o arquivo da área (ou crie `desk_<area>.go`) e um `register<Area>(mux)`; chame-o em `deskAPI.register` (`desk.go`).
2. Registre com `d.guard(admin, allowMustChange, handler)` ([backend.md](backend.md#como-uma-rota-é-escrita)). Não existe rota autenticada sem `guard`.
3. Valide **tudo** que vem do cliente (tamanhos, caminhos, ids) e filtre por `u.ID`. Erros: `deskErr(w, status, "mensagem")`.
4. Escreva o teste em `desk_<area>_test.go` com `newDK` cobrindo: sem sessão (401), usuário errado (403/404) e o caminho feliz.
5. Documente a rota na tabela da API em [desktop.md](../desktop.md#api-apidesk).

### Nova migração

1. Defina a SQL em uma constante junto da área (como `deskChatMigration`) com `CREATE TABLE IF NOT EXISTS desk_...` ou `ALTER TABLE`.
2. **Acrescente no fim** da lista `migrations` em `db.go`. **Nunca edite nem reordene** as anteriores: o `user_version` conta por posição.
3. Nomeie tabelas novas com o prefixo `desk_`: o importador da época da Central (`import_central.go`) copia por esse prefixo.
4. Teste: `openStore` em um arquivo temporário já aplica todas as migrações; confira que a tabela existe e que um banco do passo anterior sobe sem erro.

### Nova opção de configuração

1. Campo em `DeskConfig` (ou `Config`) com `json:"nome"` e comentário. Tudo é **opcional**: o valor zero precisa ser um padrão seguro.
2. Acrescente ao `diogows.config.example.json` se for útil, e à tabela do [README](../../README.md#configuração).

### Novo componente de terceiros

Copie para `www/vendor/<nome>/` com a licença, registre em `www/vendor/components.json`, no catálogo de `desk_updates.go` e em [licencas.md](../licencas.md).

## Estilo do código

- Comentários **em português**, explicando o *porquê* e as regras de segurança; cada `desk_*.go` abre com um parágrafo sobre a área.
- Sem dependências novas sem necessidade: o projeto prefere a biblioteca padrão. As que existem estão em `go.mod` e em [licencas.md](../licencas.md).
- Front-end sem framework e sem build; nunca montar HTML com texto do usuário (`DK.el`, `textContent`).
