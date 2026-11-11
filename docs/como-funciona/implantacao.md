# Empacotar e instalar

## Empacotar (no seu PC)

```bash
cd server
./package.sh            # pacote "de desenvolvimento": cada instalação soma 1 ao último número da versão
./package.sh v1.2.0     # pacote de release: instala exatamente a versão 1.2.0
```

`package.sh` chama `build.sh` e gera em `server/dist/`:

| Arquivo | Para | Conteúdo |
|---|---|---|
| `diogows.tgz` | Linux | binário de cada arquitetura (amd64, 386, arm64, arm), `install.sh`, `diogows.service`, `www.tgz` e a configuração de produção |
| `diogows-windows.zip` | Windows | `diogows-windows-amd64.exe`, `Instalar.cmd`, `install.ps1`, `www/` e o `diogows.config.example.json` |

Os binários são estáticos (`CGO_ENABLED=0`), sem dependências no sistema. A configuração de produção sai do **exemplo versionado** com `staticDir` em `./www` e a Central em `127.0.0.1:80`;
um arquivo seu com chaves só entra com `CONFIG=... ./package.sh` e **nunca deve ser publicado**.

## Instalar no Linux (o servidor da casa)

```bash
scp server/dist/diogows.tgz usuario@servidor:/tmp/
ssh usuario@servidor
mkdir /tmp/dw && tar -xzf /tmp/diogows.tgz -C /tmp/dw && cd /tmp/dw
sudo bash install.sh
```

O `install.sh`:

1. descobre a arquitetura e instala o binário em `/opt/diogows/diogows`;
2. na **primeira** vez grava `diogows.config.json`; depois o mantém (e guarda `.bak`);
3. refaz `/opt/diogows/www` e instala `diogows.service` (executa como o usuário que rodou o `sudo`, `MemoryMax=500M`, `NoNewPrivileges`);
4. libera a porta **8090** no firewall (`ufw` ou `firewalld`, se ativos);
5. `systemctl enable --now diogows`, confere o `/api/health` e grava a versão em `VERSION` / `install.log`.

Comandos úteis:

```bash
systemctl status diogows
journalctl -u diogows -f
sudo journalctl -u diogows | grep "senha inicial"     # primeira instalação
bash install.sh --check                               # versão instalada, sem alterar nada
```

### Terminal e arquivos do computador (opcional)

```bash
sudo bash install.sh --desk-shell        # sudo sem senha para o usuário do serviço (+ NoNewPrivileges=false)
sudo bash install.sh --no-desk-shell     # desfaz
```

Só vale com `"desk": {"shell": true}` no `diogows.config.json`. Leia o [SECURITY.md](../../SECURITY.md) antes: isso entrega ao Diogows o poder de root.

### Atrás de um proxy com HTTPS (opcional)

O **LiteOffice Completo** só funciona em contexto seguro (HTTPS ou `localhost`) e o login trafega em texto em HTTP. Um proxy como o nginx resolve os dois:

```nginx
server {
  listen 443 ssl;
  server_name diogows.casa;
  # ssl_certificate ... ;
  location / {
    proxy_pass http://127.0.0.1:8090;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    client_max_body_size 2g;          # o limite de envio do Diogows é desk.maxUploadMB (2 GB)
    proxy_request_buffering off;
  }
}
```

Para o retorno automático do OAuth do e-mail, `desk.mail.redirectURI` aponta para `https://diogows.casa/api/desk/mail/oauth/callback` ([email.md](../email.md)).

## Instalar no Windows

Extraia o `diogows-windows.zip` e abra o `Instalar.cmd`. A janela pergunta:

| Modo | Roda como | Escuta em | Para |
|---|---|---|---|
| **Servidor da casa** | tarefa do sistema (SYSTEM), sobe com o Windows; exige administrador | `:8090` (regra de firewall para redes privadas) | atender os outros aparelhos |
| **Meu computador** | tarefa na sua sessão | `127.0.0.1:8090` | só você |

Ao terminar, a janela mostra o endereço e a **senha inicial** do `admin` (anote: ela não aparece de novo). Atualizar = extrair a versão nova e abrir o `Instalar.cmd` de novo; os dados são mantidos.
Desinstalar: *Aplicativos instalados* do Windows ou `Desinstalar.cmd` na pasta de instalação (pergunta se apaga os dados). Parâmetros para automação: `install.ps1 -Modo servidor|pessoal -Silencioso -Pasta <dir> -Porta <n>`.

No Windows o Terminal com `sudo` não existe, e o log vai para `diogows.log` na pasta de instalação.

## Versões

Política em [versionamento.md](../versionamento.md): tags `vMAIOR.MENOR.PATCH` saindo da `main`. O workflow `.github/workflows/release.yml` roda `go vet`, `go test` e `check-i18n`,
empacota, gera o `SHA256SUMS` e cria a release. A versão instalada aparece em `GET /api/health` e no aplicativo Computador.

## Quando o Diogows e a Central estão no mesmo servidor

| | Central Diogo | Diogows | Navidrome |
|---|---|---|---|
| Pasta | `/opt/centraldiogo` | `/opt/diogows` | (dele) |
| Serviço | `centraldiogo-hub` | `diogows` | `navidrome` |
| Porta | 80 | 8090 | 4533 |
| Banco | `centraldiogo.db` | `diogows.db` | (dele) |

Cada um instala, atualiza, reinicia e faz backup **sem afetar os outros**. O Diogows chama a Central em `http://127.0.0.1:80`; se ela cair, só os itens que dependem dela ficam indisponíveis.
