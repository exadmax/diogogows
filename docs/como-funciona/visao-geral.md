# Visão geral

## O que é

O Diogows é uma **área de trabalho no navegador** para a casa: janelas, menu Iniciar, ícones, barra de tarefas e dezenas de aplicativos
(Explorer, documentos, planilhas, apresentações, e-mail, mensagens, calendário, navegador, Paint, player de música, emuladores, programas DOS...).
Cada pessoa da casa tem **login** e a **própria pasta** de arquivos; o administrador decide quem pode usar o Terminal e os arquivos do computador.

Roda em qualquer navegador (PC, tablet, celular, TV) e o servidor cabe em uma máquina de 2 GB de RAM, dividindo-a com outros serviços.

## O que ele não é

- **Não é a Central Diogo.** A Central é o *painel da casa* (agenda, clima, trânsito, câmeras ao vivo, estoque da cozinha, livros). É outro programa, em outro
  repositório e outro serviço; o Diogows a usa só como **fonte** de câmeras, contatos e agendas (veja [central.md](central.md)).
- **Não é um sistema operacional nem uma máquina virtual.** As "janelas" são `<div>`s numa página; os programas DOS rodam no navegador (js-dos) e o LibreOffice,
  em WebAssembly (ZetaOffice). O servidor só guarda arquivos e faz o que o navegador não pode (e-mail, buscar páginas, terminal).
- **Não é para a internet.** É feito para uma rede doméstica confiável (veja [seguranca.md](seguranca.md)).

## Decisões que moldam o projeto

| Decisão | Por quê | Consequência |
|---|---|---|
| **Um binário Go, sem CGO** | Instalar = copiar um arquivo; compila para Linux (amd64, 386, arm64, arm) e Windows | SQLite em Go puro (`modernc.org/sqlite`); sem bibliotecas de sistema |
| **Front-end em JavaScript puro, sem build** | O que está em `www/` é exatamente o que roda; sem Node, sem bundler, depuração direta no navegador | Cada aplicativo é um arquivo que se registra em `DK`; carregamento sob demanda feito à mão (`lazy.js`) |
| **SQLite + pasta de arquivos** | Poucos usuários, uma máquina; backup é copiar uma pasta | Uma conexão de escrita e um pequeno pool de leitura (`db.go`) |
| **Sessão em cookie, senha com PBKDF2** | Simples e auditável; nenhum serviço externo | Login por usuário, bloqueio por tentativas, troca de senha obrigatória no primeiro acesso |
| **O servidor faz o que o navegador não pode** | CORS, segredos e rede interna | Navegador interno, e-mail (IMAP/SMTP), downloads, ícones de sites, agendas iCal passam pelo servidor, com filtros de segurança |
| **Separado da Central** | Cresceram tanto que cada um merece ciclo de versões próprio | Ligação por HTTP (`central.url`), sem banco compartilhado; funciona sozinho |
| **Tudo opcional e desligável** | Cada casa usa um pedaço | `desk.shell`, `desk.noBrowser`, `desk.mail.disabled`, `desk.updates.disabled`... |

## Vocabulário

| Termo | Significado |
|---|---|
| **desk** | Prefixo histórico de tudo do Diogows no servidor: `/api/desk/*`, `desk_*.go`, tabelas `desk_*`, a pasta `desk/`. Veio da época em que se chamava "Desktop" |
| **`DK`** | O objeto global do front-end (`window.DK`): janelas, API, aplicativos registrados |
| **Meus arquivos** (`space=home`) | A pasta própria de cada usuário no servidor (`desk/<id>/files`) |
| **Computador** (`space=machine`) | O sistema de arquivos do servidor, só com permissão do administrador e `desk.shell` ligado |
| **Central** | A Central Diogo (painel da casa), outro programa |
| **Admin** | Usuário com `is_admin`: gerencia usuários, câmeras da Central, atualizações e vê tudo |
