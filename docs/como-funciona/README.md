# Como o Diogows funciona

Esta pasta explica o projeto por dentro: o que ele é, como as peças se encaixam, onde fica cada coisa no código e como fazer as mudanças
mais comuns. Quem só quer **usar** deve ler antes o [README](../../README.md) e o [guia do desktop](../desktop.md).

| Leia | Para entender |
|---|---|
| [visao-geral.md](visao-geral.md) | O que é o Diogows, o que ele **não** é, e as decisões que moldam o resto |
| [arquitetura.md](arquitetura.md) | Servidor, front-end, como uma janela conversa com a API, e o que roda onde |
| [backend.md](backend.md) | O servidor em Go: arquivos, rotas, sessão, permissões, banco, arquivos dos usuários |
| [frontend.md](frontend.md) | O front-end: `DK`, janelas, aplicativos, carregamento sob demanda, páginas isoladas |
| [central.md](central.md) | A ligação com a Central Diogo: o que é repassado, por que e com quais regras |
| [dados.md](dados.md) | Onde ficam banco, arquivos, chaves e configuração, e o que entra no backup |
| [seguranca.md](seguranca.md) | O modelo de segurança em uma página, e onde cada regra é aplicada no código |
| [desenvolvimento.md](desenvolvimento.md) | Rodar no seu computador, testar, e receitas: novo aplicativo, nova rota, nova migração |
| [implantacao.md](implantacao.md) | Empacotar, instalar no Linux e no Windows, versões e atualização |
| [migracao-da-central.md](migracao-da-central.md) | Para quem usava o Diogows quando ele morava dentro da Central Diogo |

## Em uma frase

Um único programa em **Go** (`server/`) que serve um front-end em **JavaScript puro, sem build** (`www/`) e uma API `/api/desk/*` com login e vários
usuários; ele guarda tudo em **SQLite** e numa pasta de arquivos, e se liga, opcionalmente, à **Central Diogo** para buscar câmeras, contatos e agendas da casa.

## Mapa do repositório

```text
diogows/
├── README.md                 apresentação, instalação rápida, configuração
├── SECURITY.md               modelo de ameaça e proteções
├── docs/                     guias de uso (desktop.md, email.md, ...) e esta pasta (como-funciona/)
├── scripts/                  check-i18n.js (confere as traduções)
├── server/                   o programa em Go (módulo "diogows")
│   ├── main.go               liga tudo e sobe o servidor
│   ├── config.go             diogows.config.json
│   ├── db.go                 SQLite, migrações, respostas JSON
│   ├── central.go            ponte para a Central Diogo
│   ├── desk.go               usuários, sessões, login, portões de permissão
│   ├── desk_*.go             um arquivo por área: arquivos, terminal, e-mail, chat, ...
│   ├── games*.go             placar de jogos (Steam, Xbox, PlayStation, RetroAchievements)
│   ├── static.go, office.go  entrega do front-end (gzip, cache, isolamento do LiteOffice)
│   ├── import_central.go     importa os dados da época em que morava na Central
│   ├── install.sh, package.sh, build.sh, diogows.service, windows/   empacotamento e instalação
│   └── demo/                 instância de demonstração (dados ignorados pelo git)
└── www/                      o front-end (o que o navegador baixa)
    ├── desktop.html          a área de trabalho (index.html só redireciona para ela)
    ├── office.html           página própria do LiteOffice Completo (precisa de cabeçalhos especiais)
    ├── emulator.html, emu.html   páginas dos emuladores (DOS e consoles), cada jogo num iframe
    ├── js/desktop/           um arquivo por aplicativo, mais o núcleo (core.js, boot.js, shell.js)
    ├── js/core, js/services, js/i18n   utilitários e traduções (cópia enxuta da base da Central)
    ├── css/desktop.css       todos os estilos
    └── vendor/               bibliotecas de terceiros (JS Paint, Webamp, js-dos, ZetaJS)
```
