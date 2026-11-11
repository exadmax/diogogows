# Vindo da Central Diogo

Até esta versão o Diogows morava **dentro** da Central Diogo: um único programa (`centraldiogo-hub`), um banco (`centraldiogo.db`) e uma pasta de dados (`desk/`). Agora são dois programas.
Esta página é o passo a passo para quem já usava o Diogows e quer **levar usuários, mensagens, e-mails e arquivos** para o novo serviço, sem perder nada.

## O que muda

| | Antes | Agora |
|---|---|---|
| Programa | `centraldiogo-hub` (um só) | `centraldiogo-hub` (Central) **+** `diogows` |
| Pasta | `/opt/centraldiogo` | `/opt/centraldiogo` **e** `/opt/diogows` |
| Endereço do Diogows | `http://servidor/desktop.html` | `http://servidor:8090/` (o endereço antigo **redireciona**) |
| Banco do Diogows | tabelas `desk_*` dentro de `centraldiogo.db` | `diogows.db` próprio |
| Arquivos dos usuários | `/opt/centraldiogo/desk/` | `/opt/diogows/desk/` |
| Config | bloco `desk`, `games` e `videos` do `hub.config.json` | `diogows.config.json` |
| `protectSettings` | existia | **removida** (a Central continua sem login) |
| Jogos (`games`) | cartão na Config da Central | só no `diogows.config.json` |

O que **continua na Central**: câmeras, agendas, trânsito, estoque, profissionais, livros, vídeos, gravações, o menu e a Config.

## Ordem recomendada

Instale o **Diogows primeiro** e só depois atualize a Central: assim o desktop nunca fica fora do ar.

### 1. Instalar o Diogows importando da Central

No servidor, com o pacote do Diogows extraído:

```bash
sudo bash install.sh --import-central /opt/centraldiogo
```

O instalador instala o serviço **e, antes de iniciá-lo**, importa a partir da pasta da Central (parando a Central por alguns segundos para ler o banco dela de forma consistente, e religando-a em seguida):

| Importado | De onde | Para onde |
|---|---|---|
| Usuários, sessões, agendas, mensagens, contas de e-mail, perfis, planta da casa | tabelas `desk_*` do `centraldiogo.db` | `diogows.db` |
| Arquivos dos usuários, chave do e-mail (`_mail.key`), anexos do chat | `/opt/centraldiogo/desk/` (**copiados**) | `/opt/diogows/desk/` |
| Blocos `desk`, `games` e `videos` | `hub.config.json` | `diogows.config.json` (sem `protectSettings`) |
| `games.state.json` | pasta da Central | pasta do Diogows |
| `central.url` | a porta em que a Central escuta | `diogows.config.json` |

**Nada da Central é alterado ou apagado.** A importação só roda se o `diogows.db` ainda não existir (e recusa um banco que já tenha usuários).
Se algo falhar, o instalador avisa e a Central volta a rodar como estava.

### 2. Conferir

Abra `http://servidor:8090/`, entre com um usuário que já existia (as **mesmas senhas**), abra *Meus arquivos* e o *E-mail*. Confira também Câmeras e Contatos (vêm da Central).

### 3. Atualizar a Central

Instale a nova versão da Central normalmente (`sudo bash install.sh`). Ela já não traz o desktop; quem abrir `http://servidor/desktop.html` é **redirecionado** para o Diogows.
O atalho "Diogows" do menu da Config (que apontava para `/desktop.html`) continua funcionando pelo redirecionamento; para deixá-lo direto, edite-o em *Config → Menu*:
porta `8090`, caminho `/`.

### 4. Limpar (quando tiver certeza)

Depois de alguns dias de uso, se tudo está certo, libere espaço:

```bash
sudo rm -r /opt/centraldiogo/desk        # os arquivos que agora vivem em /opt/diogows/desk
```

As tabelas `desk_*` antigas ficam no `centraldiogo.db`; são inofensivas e ocupam pouco. Não há necessidade de apagá-las.

## Voltar atrás

Como nada foi apagado, voltar é simples: `sudo systemctl stop diogows`, reinstale a versão antiga da Central (ela tem o desktop e usa as tabelas `desk_*` e a `desk/` que ficaram lá).
Mudanças feitas **depois** da migração (mensagens novas, arquivos novos) ficam só no Diogows.

## Importar à mão

O importador é o próprio programa: com o serviço **parado**,

```bash
/opt/diogows/diogows -config /opt/diogows/diogows.config.json -import-central /opt/centraldiogo
# ou só o banco:
/opt/diogows/diogows -config /opt/diogows/diogows.config.json -import-central /caminho/centraldiogo.db
```

Com uma **pasta** ele faz tudo da tabela acima; com um **arquivo** `.db`, só as tabelas `desk_*`. A saída mostra quantas linhas de cada tabela foram copiadas.
No Windows copie os dados à mão (a janela do instalador não faz a importação) ou rode o comando acima no PowerShell.

## Se o Diogows já estava em uso na Central e você não quer migrar

Pode simplesmente começar do zero: instale o Diogows sem `--import-central`. Ele cria um `admin` novo (senha no log).
