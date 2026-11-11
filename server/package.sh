#!/usr/bin/env bash
# Gera dist/diogows.tgz (Linux) e dist/diogows-windows.zip (Windows) com tudo que cada sistema precisa (roda no seu PC).
# Uso: ./package.sh [vX.Y.Z]   (com a versao, o pacote vira uma release: o install.sh grava essa versao em VERSION)
set -e
VER="${1#v}"
cd "$(dirname "$0")"
./build.sh >/dev/null
P=dist/pkg; rm -rf $P; mkdir -p $P
cp dist/diogows-linux-* install.sh diogows.service $P/
# config de producao: o front-end fica em ./www e a Central escuta na porta 80 (o arquivo de origem nao e alterado).
sed -e 's#"staticDir": *"[^"]*"#"staticDir": "./www"#' -e 's#"url": *"http://127.0.0.1:8080"#"url": "http://127.0.0.1:80"#' diogows.config.example.json > $P/diogows.config.json
[ -n "$VER" ] && echo "$VER" > $P/RELEASE
tar -C ../www -czf $P/www.tgz .
tar -C $P -czf dist/diogows.tgz .

# Windows: zip com o instalador em janela (windows/), o programa, o front-end em ./www e a configuracao de exemplo
W=dist/winpkg; rm -rf $W; mkdir -p $W/www
cp dist/diogows-windows-amd64.exe windows/install.ps1 windows/Instalar.cmd windows/LEIA-ME.txt diogows.config.example.json $W/
[ -n "$VER" ] && echo "$VER" > $W/RELEASE
tar -C ../www -cf - . | tar -C $W/www -xf -
PY=python3; command -v $PY >/dev/null || PY=python
rm -f dist/diogows-windows.zip
(cd $W && $PY -m zipfile -c ../diogows-windows.zip *)
ls -lh dist/diogows.tgz dist/diogows-windows.zip
