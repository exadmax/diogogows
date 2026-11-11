# Segurança

## Modelo de ameaça
O Diogows é feito para uma **rede doméstica confiável**, com **login e vários usuários**. Ele protege o que é *dele* (arquivos dos usuários,
agendas, e-mail, mensagens, Terminal), mas continua sendo um programa de casa: **não o exponha à internet** (nem por redirecionamento de porta).
Para acesso remoto use uma VPN (Tailscale, WireGuard) ou um proxy reverso com HTTPS e autenticação na frente.

- Em **HTTP puro** (o normal na rede de casa) as senhas trafegam **sem criptografia**. Fora de casa, só por VPN ou HTTPS.
- A configuração (`diogows.config.json`) pode ter chaves de API e senhas em texto, e o banco guarda dados pessoais: o instalador cria os arquivos com permissão restrita ao dono.
  Não os versione nem os compartilhe; o `.gitignore` já os exclui.

## A ligação com a Central Diogo
A [Central Diogo](https://github.com/exadmax/DVRCentral) (o painel da casa) é **outro programa e não tem login**. O Diogows só repassa, **para quem está logado**, as rotas dela
que as janelas usam (`/api/cameras`, `/api/contatos`, `/api/ics` e `/api/settings/*`, veja [docs/como-funciona/central.md](docs/como-funciona/central.md)):
leitura exige sessão; escrita exige sessão e origem igual à do Diogows; alterar `/api/settings/*` exige **administrador**. O cookie de sessão nunca vai para a Central.

Isso protege o caminho **pelo Diogows**, não a Central em si: quem alcança a porta da Central continua vendo as câmeras e a Config dela. Se isso importa, deixe a Central
escutando só em `127.0.0.1` (ou atrás de um firewall) e acesse tudo pelo Diogows. A opção `protectSettings`, que existia quando os dois eram um programa só, foi removida.

## Login, usuários e terminal
- **Terminal e arquivos do computador** estão **desligados por padrão** (`desk.shell`). Ligados, dão a quem tem permissão o poder do usuário do serviço;
  com `install.sh --desk-shell`, o **poder de root** (sudoers NOPASSWD). A senha de um administrador do Diogows passa a ser, na prática, a senha de root.
  Cada uso exige: recurso ligado, permissão do administrador, rede local ou HTTPS, a própria senha (15 min) e, para o sudo, a senha de um administrador. Tudo vai para o journal.
- Em **HTTP puro** (o normal na rede de casa) as senhas do Diogows trafegam **sem criptografia**. Use só em rede confiável; fora de casa, VPN ou HTTPS.
- A senha inicial do `admin` é **sorteada** e aparece uma única vez no log; a troca é obrigatória. Senhas ficam com PBKDF2-SHA256 (600 mil iterações); o login tem bloqueio por tentativas.
- As **agendas iCal** de cada usuário são segredos (quem tem a URL lê a agenda): ficam no banco, saem mascaradas pela API e só o dono as vê. O servidor só baixa de endereços
  públicos (não de rede interna nem de si mesmo), para o campo de URL não servir de sonda para a sua rede.
- O **Navegador** busca as páginas pelo servidor e as entrega sem scripts (limpeza do HTML, política de segurança e iframe sem JavaScript). Só alcança **sites da internet**: endereços
  da rede da casa, do próprio servidor e de nuvem (metadados) são recusados. Os downloads vão para a pasta do usuário e nada baixado é executado no servidor. `desk.noBrowser` desliga tudo isso.
- Os **emuladores** (js-dos) rodam no navegador, num iframe isolado, sem rede (`noNetworking`) e sem a nuvem do js-dos. O Windows (3.x, 95, 98, ME) é software proprietário e **não é distribuído**: use só imagens que você tenha licença para usar. O disco alterado de um sistema virtual fica na pasta do próprio usuário no servidor (`Programas/Sistemas`), com o mesmo isolamento dos outros arquivos.
- O editor de documentos e o visualizador nunca interpretam HTML, SVG ou scripts de arquivos: texto, imagem, PDF e vídeo são servidos com `nosniff` e política restritiva.
- O **LiteOffice Completo** (`office.html`, LibreOffice em WebAssembly) executa, no navegador, código baixado do CDN do ZetaOffice (terceira parte) ou da cópia local configurada. O endereço de onde esse código vem só é lido da configuração (`www/config/settings.js`), nunca da URL da página, e os cabeçalhos de isolamento (COOP/COEP) valem só para essa página e seus scripts, não para o resto do Diogows.
- O **E-mail** guarda senhas e *refresh tokens* **cifrados** (AES-GCM) no banco, com a chave em `desk/_mail.key` (modo 0600; faça backup dela, e lembre que quem lê o disco do servidor lê a chave junto: a cifra protege cópias do banco, não um servidor comprometido).
  Só fala TLS (implícito ou STARTTLS), só conecta em endereços da internet (a menos que `desk.mail.allowPrivate`), cada usuário só alcança as próprias contas e a API nunca devolve credenciais. O HTML das mensagens roda num iframe com `sandbox`, sem scripts e sem rede;
  **imagens remotas só carregam se o usuário pedir**; anexos saem sempre como download (nunca HTML ou SVG). O login OAuth usa PKCE e `state` de uso único. Veja [docs/email.md](docs/email.md#7-segurança).
- O **perfil** (foto, nome, assinatura): a foto só é aceita se for um JPEG ou PNG **decodificável** (até 400 KB), e é servida como imagem inerte (`nosniff`, CSP restritiva). As fotos são visíveis a qualquer usuário logado.
- A busca do **ícone de um site** (Programa Web) é feita pelo servidor, só em endereços da internet (mesmo bloqueio do Navegador interno), aceita só imagens reconhecidas pelo conteúdo e limita o tamanho. `desk.noBrowser` a desliga.
- Um **Programa Web** na janela interna carrega a página sem `sandbox`: ela tem os poderes de uma aba comum. Cadastre só endereços em que confia. O WhatsApp Web e outros sites que proíbem ser exibidos dentro de outra página abrem em janela própria do navegador; o Diogows não contorna essas proteções.
- A tela **Atualizações** (só administrador) consulta versões em três endereços fixos (npm, GitHub e proxy do Go), só leitura; o Webamp baixado tem o hash sha512 conferido ([docs/atualizacoes.md](docs/atualizacoes.md)).
- O **Paint** (JS Paint) e o **Player de Música** (Webamp) são cópias de projetos MIT dentro do Diogows ([docs/licencas.md](docs/licencas.md)). O Paint roda num iframe do mesmo servidor com **política de conteúdo fechada** (sem Firebase, YouTube, fontes externas nem imagens de outros sites); o Webamp é carregado direto na página do Diogows (sem iframe, mas sem usar a rede nem atalhos globais de teclado). Skins `.wsz` são arquivos de terceiros: use só os de confiança.

## Reportar uma vulnerabilidade
Abra um *security advisory* privado no GitHub (aba **Security** do repositório) em vez de uma issue pública. Descreva o problema e
como reproduzi-lo.
