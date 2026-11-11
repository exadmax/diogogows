#!/usr/bin/env bash
# Roda NO Linux, com sudo:  sudo bash install.sh
# Instala o Diogows em /opt/diogows (porta 8090) e ativa o serviço "diogows", ao lado da Central Diogo, do Navidrome e dos demais.
# Pacote de release (tem o arquivo RELEASE, gravado por package.sh a partir da tag vX.Y.Z): a instalação grava essa versão em VERSION.
# Pacote sem RELEASE (montado do código, sem tag): cada instalação concluída soma 1 ao terceiro número (VERSION: 1.0.1 -> 1.0.2).
# Só consultar a versão instalada, sem sudo e sem alterar nada:  bash install.sh --check
#
# Vindo da Central Diogo (o Diogows já morou dentro dela): leva usuários, mensagens, e-mails, arquivos e configuração para cá.
#   sudo bash install.sh --import-central /opt/centraldiogo
# Só funciona na primeira instalação (banco do Diogows ainda inexistente) e não altera nem apaga nada da Central.
#
# Terminal e arquivos do computador no Diogows (opcional; veja SECURITY.md):
#   sudo bash install.sh --desk-shell        permite ao serviço usar sudo (sudoers NOPASSWD + NoNewPrivileges=false)
#   sudo bash install.sh --no-desk-shell     desfaz isso
set -e
cd "$(dirname "$0")"
D=/opt/diogows
PORT=8090

# Versão instalada = arquivo VERSION ("MAIOR.MENOR.PATCH"). Sem o arquivo ainda não há instalação versionada: parte de 1.0.0.
# Para subir o MAIOR ou o MENOR, edite o VERSION à mão (ex.: 1.1.0); as instalações seguintes continuam somando no terceiro.
OLD=1.0.0
[ -f "$D/VERSION" ] && OLD="$(tr -d '[:space:]' < "$D/VERSION")"
if ! echo "$OLD" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$'; then
  echo "!! $D/VERSION inválido: '$OLD' (esperado algo como 1.0.1). Corrija o arquivo e rode de novo."; exit 1
fi
BASE="${OLD%%-*}"   # sem o sufixo de pré-lançamento (-rc.1)
NEW="${BASE%.*}.$((10#${BASE##*.} + 1))"
# Versão da release empacotada (ver docs/versionamento.md): manda no número instalado.
if [ -f RELEASE ]; then
  REL="$(tr -d '[:space:]' < RELEASE)"
  echo "$REL" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$' || { echo "!! RELEASE inválido: '$REL'"; exit 1; }
  NEW="$REL"
fi

if [ "$1" = "--check" ]; then
  if [ -f "$D/VERSION" ]; then echo "versão instalada: V$OLD (desde $(date -r "$D/VERSION" '+%F %T'))"
  else echo "versão instalada: nenhuma registrada ainda"; fi
  if [ -f RELEASE ]; then echo "este pacote instala: V$NEW"; else echo "a próxima instalação será: V$NEW"; fi
  exit 0
fi

[ "$(id -u)" = 0 ] || { echo "rode com sudo: sudo bash install.sh"; exit 1; }
USER_RUN="${SUDO_USER:-root}"
DESK_SHELL=""
IMPORT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --desk-shell) DESK_SHELL=on;;
    --no-desk-shell) DESK_SHELL=off;;
    --import-central) shift; IMPORT="${1:-}"; [ -n "$IMPORT" ] || { echo "!! --import-central precisa da pasta da Central (ex.: /opt/centraldiogo)"; exit 1; };;
  esac
  shift
done

case "$(uname -m)" in
  x86_64) A=amd64;; i?86) A=386;; aarch64|arm64) A=arm64;; armv6*|armv7*|arm) A=arm;;
  *) echo "arquitetura não suportada: $(uname -m)"; exit 1;;
esac
echo ">> arquitetura: $A"

# a porta do Diogows precisa estar livre (ou já ser dele, numa atualização)
if ss -ltn 2>/dev/null | awk '{print $4}' | grep -qE "[:.]$PORT\$" && ! systemctl is-active --quiet diogows; then
  echo "!! Já existe algo escutando na porta $PORT:"; ss -ltnp | grep -E "[:.]$PORT " || true
  echo "   Mude \"listen\" em $D/diogows.config.json (e a porta em install.sh) ou libere a porta."; exit 1
fi

mkdir -p "$D/www"
systemctl stop diogows 2>/dev/null || true
install -m 755 "diogows-linux-$A" "$D/diogows"
# A configuração (diogows.config.json) é preservada: o pacote só a instala na primeira instalação.
# Para trocar algo depois, edite o arquivo no servidor.
FIRST=""
if [ -f "$D/diogows.config.json" ]; then
  cp "$D/diogows.config.json" "$D/diogows.config.json.bak"
  echo ">> diogows.config.json existente mantido (cópia em diogows.config.json.bak)"
else
  install -m 600 diogows.config.json "$D/diogows.config.json"
  FIRST=1
fi
rm -rf "$D/www" && mkdir "$D/www" && tar -xzf www.tgz -C "$D/www"
chown -R "$USER_RUN" "$D"
sed "s/__USER__/$USER_RUN/" diogows.service > /etc/systemd/system/diogows.service

# Vindo da Central: copia banco, arquivos e configuração (a Central fica intacta). Só se o banco do Diogows ainda não existe.
if [ -n "$IMPORT" ]; then
  if [ -f "$D/diogows.db" ]; then
    echo ">> o Diogows já tem banco ($D/diogows.db): importação da Central ignorada"
  else
    [ -d "$IMPORT" ] || { echo "!! pasta da Central não encontrada: $IMPORT"; exit 1; }
    echo ">> importando de $IMPORT (a Central não é alterada) ..."
    # com o serviço da Central parado o banco dela fica consistente (WAL fechado) durante a cópia
    CENTRAL_WAS_ACTIVE=""
    if systemctl is-active --quiet centraldiogo-hub; then CENTRAL_WAS_ACTIVE=1; systemctl stop centraldiogo-hub; fi
    sudo -u "$USER_RUN" "$D/diogows" -config "$D/diogows.config.json" -import-central "$IMPORT" \
      || { [ -n "$CENTRAL_WAS_ACTIVE" ] && systemctl start centraldiogo-hub; echo "!! a importação falhou; nada foi apagado"; exit 1; }
    [ -n "$CENTRAL_WAS_ACTIVE" ] && systemctl start centraldiogo-hub
  fi
fi

# Sudo do usuário do serviço, para o Terminal e os arquivos do computador. É opt-in porque dá ao Diogows o poder de root:
# quem tiver a permissão (e a senha de um administrador do Diogows) passa a administrar a máquina. Só vale com "desk": {"shell": true}
# no diogows.config.json. O sudo dentro de um serviço com NoNewPrivileges=true não funciona, por isso o drop-in abaixo.
DROPIN=/etc/systemd/system/diogows.service.d/desk-shell.conf
SUDOERS=/etc/sudoers.d/diogows-desk
if [ "$DESK_SHELL" = on ]; then
  [ "$USER_RUN" = root ] && { echo "!! --desk-shell precisa do usuário normal (rode com sudo a partir dele, não como root direto)"; exit 1; }
  mkdir -p "$(dirname "$DROPIN")"
  printf '%s\n' '[Service]' 'NoNewPrivileges=false' > "$DROPIN"
  TMP="$(mktemp)"; echo "$USER_RUN ALL=(ALL) NOPASSWD: ALL" > "$TMP"
  visudo -cf "$TMP" >/dev/null || { rm -f "$TMP"; echo "!! sudoers inválido; nada foi alterado"; exit 1; }
  install -m 440 "$TMP" "$SUDOERS"; rm -f "$TMP"
  echo ">> sudo liberado para $USER_RUN (arquivos: $DROPIN e $SUDOERS)"
  echo ">> falta ligar no diogows.config.json:  \"desk\": {\"shell\": true}   e depois: sudo systemctl restart diogows"
elif [ "$DESK_SHELL" = off ]; then
  rm -f "$DROPIN" "$SUDOERS"; rmdir "$(dirname "$DROPIN")" 2>/dev/null || true
  echo ">> sudo do Diogows desfeito"
fi

# firewall: libera a porta se houver ufw/firewalld ativo
if command -v ufw >/dev/null && ufw status | grep -q "Status: active"; then ufw allow $PORT/tcp; echo ">> ufw: porta $PORT liberada"
elif command -v firewall-cmd >/dev/null && systemctl is-active --quiet firewalld; then firewall-cmd --permanent --add-port=$PORT/tcp && firewall-cmd --reload; echo ">> firewalld: porta $PORT liberada"
else echo ">> sem firewall ativo detectado (ufw/firewalld); nada a liberar"; fi

systemctl daemon-reload
systemctl enable diogows
systemctl restart diogows
sleep 1.5
systemctl is-active diogows

# instalação concluída (o serviço subiu): registra a nova versão. O Diogows lê o arquivo, então /api/health já mostra a nova.
echo "$NEW" > "$D/VERSION.tmp" && mv "$D/VERSION.tmp" "$D/VERSION"
echo "$(date '+%F %T')  V$OLD -> V$NEW" >> "$D/install.log"
chown "$USER_RUN" "$D/VERSION" "$D/install.log"
echo ">> versão instalada: V$NEW (era V$OLD)"
curl -s "localhost:$PORT/api/health" && echo && echo ">> pronto: http://$(hostname -I | awk '{print $1}'):$PORT/desktop.html"
if [ -n "$FIRST" ] && [ -z "$IMPORT" ]; then
  echo ">> primeira instalação: a senha inicial do usuário admin está no log:  sudo journalctl -u diogows | grep 'senha inicial'"
fi
