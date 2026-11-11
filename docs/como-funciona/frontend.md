# O front-end

Tudo em `www/`. **JavaScript puro, sem build, sem framework**: o que está no disco é o que o navegador executa. Isso é deliberado (depuração direta, nada para instalar),
e tem um custo que o projeto aceita: os módulos se comunicam por um objeto global, `DK`.

## O objeto `DK`

Cada arquivo de `js/desktop/` é uma função autoexecutável que recebe `DK` e acrescenta o que oferece:

```js
(function (DK) {
  const el = DK.el;
  DK.contactsApp = function () { /* abre a janela */ };
  DK.register({ id: 'contacts', icon: '📇', label: 'Contatos', desktop: true, open: DK.contactsApp });
})(window.DK = window.DK || {});
```

| Peça | Onde | Para quê |
|---|---|---|
| `DK.el(tag, props, ...filhos)`, `DK.$` | `core.js` | Criar elementos sem HTML em texto (evita injeção) |
| `DK.api(método, caminho, corpo)` | `core.js` | Chamar `/api/desk/*` (envia `X-Desk: 1`, trata erros como `Error` com `.status`) |
| `DK.hub(método, caminho, corpo)` | `core.js` | Chamar rotas **da Central** repassadas pelo Diogows (`/api/cameras`, `/api/contatos`, ...) |
| `DK.centralUrl(caminho)` | `core.js` | Endereço de uma página da Central (para links e iframes), a partir de `config/settings.js` |
| `DK.fileUrl(item, modo)` | `core.js` | URL de um arquivo de *Meus arquivos* (`/api/desk/fs/read?...`) |
| `DK.openWindow(id, título, construtor, largura, altura, opções)` | `core.js` | Cria (ou foca) uma janela; o construtor recebe o `body` |
| `DK.register(app)` / `DK.openApp(id)` | `core.js` | Registrar um aplicativo no menu e nos ícones / abri-lo |
| `DK.me`, `DK.prefs`, `DK.savePrefs()` | `core.js`, `boot.js` | Usuário logado e preferências da conta (guardadas no servidor: `desk_prefs`) |

## Quem faz o quê

| Arquivo(s) | Responsabilidade |
|---|---|
| `layout.js` | Escolhe o modo de tela (computador × celular/TV) antes de tudo |
| `core.js` | `DK`: API, janelas (arrastar, redimensionar, minimizar), utilitários |
| `boot.js` | Login, troca obrigatória de senha, menu Iniciar (categorias, mais usados, busca), ícones da área de trabalho, relógio, atalhos do painel |
| `shell.js` | Barra de tarefas, menu de contexto, toque longo |
| `theme.js` | Aparência (fundo, cores do menu e da barra) |
| `lazy.js` | **Carregamento sob demanda** (abaixo) |
| `files.js`, `viewers.js`, `video.js`, `zip.js` | Explorer, visualizadores (texto, imagem, PDF), player de vídeo, compactados |
| `docs.js`, `sheets.js`, `slides.js` (+ `docfmt`, `sheetfmt`, `slidefmt`, `formula`) | Documentos, planilhas e apresentações (LiteOffice) |
| `office.js` | Abre o LiteOffice Completo em `office.html` |
| `mail.js`, `messages.js`, `calendar.js`, `contacts.js` | E-mail, mensagens, calendário, contatos |
| `browser.js`, `programs.js`, `emulators.js`, `paint.js`, `music.js`, `games.js` | Navegador, programas DOS/Windows/web, emuladores, Paint, Webamp, jogos clássicos |
| `cameras.js`, `home.js` | Câmeras (da Central) e planta da casa |
| `terminal.js`, `computer.js` | Terminal e informações do computador |
| `admin.js`, `panel.js`, `updates.js` | Usuários, Painel de Controle, Atualizações |

## Carregamento sob demanda (`lazy.js`)

Os aplicativos pesados (editores, e-mail, navegador, Paint...) **não** vêm no `desktop.html`: o `lazy.js` registra o aplicativo no menu e deixa uma "ponte" no lugar da função de abrir.
No primeiro clique, a ponte baixa o grupo de scripts, e repete a chamada na função de verdade. O Diogows abre bem mais rápido em Wi-Fi fraco, TV e computador antigo.
Para tornar outro aplicativo sob demanda: tire o `<script>` do `desktop.html`, acrescente o grupo em `GROUPS` e leve o `DK.register` dele para o `lazy.js`.

## Páginas isoladas

Três coisas não rodam na página principal, por exigirem isolamento:

| Página | Por quê separada | Como é aberta |
|---|---|---|
| `office.html` | O LibreOffice em WebAssembly precisa de `SharedArrayBuffer` e, por isso, dos cabeçalhos COOP/COEP (`office.go`) | Nova aba, com `?space=&path=` |
| `emulator.html` | O js-dos guarda estado global e só inicia limpo num documento novo; roda sem rede (`noNetworking`) | `<iframe>` dentro da janela do programa |
| `emu.html` | Idem, para o EmulatorJS (consoles) | `<iframe>` dentro da janela do jogo |

O **JS Paint** (`vendor/jspaint/`) também vai num iframe, do mesmo servidor, com política de conteúdo fechada.

## Estilos

`css/desktop.css` é o único arquivo de estilos. Tema por variáveis CSS (a janela *Personalizar* as altera). Em telas estreitas as janelas abrem em tela cheia e os alvos de toque aumentam.

## Configuração do navegador (`config/settings.js`)

`config/settings.example.js` define `window.CD_SETTINGS` com os padrões (endereço da Central, atalhos de reserva, origem do ZetaOffice). Quem quer mudar cria `config/settings.js`
(ignorado pelo git) com `window.CD_SETTINGS_LOCAL = { ... }`, mesclado por cima. O instalador do Windows e o do Linux gravam a porta da Central ali quando ela não é a padrão.

## Traduções

A interface das janelas do Diogows é **só em português**, escrita direto nos arquivos. `js/core/i18n.js` e `js/i18n/*.js` (pt, en, es, zh) existem para os módulos
herdados da Central (`services/calendar.js`, `services/cameraPlayer.js`); `node scripts/check-i18n.js` confere que continuam consistentes.

## Segurança no front-end

- Nunca montar HTML com texto de usuário: usar `DK.el` / `textContent`. O HTML de e-mail e do navegador interno vai para `iframe` com `sandbox`/`srcdoc` já limpo pelo servidor.
- Escritas sempre por `DK.api` (cabeçalho `X-Desk`); o servidor recusa escrita sem ele.
- Endereços de arquivo usam `encodeURIComponent` em `path`; o servidor revalida tudo.
