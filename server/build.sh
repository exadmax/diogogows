#!/usr/bin/env bash
# Compila o Diogows para Linux (4 arquiteturas) e Windows (amd64), binarios estaticos sem dependencias. Uso: ./build.sh
set -e
cd "$(dirname "$0")"
mkdir -p dist
for t in amd64 386 arm64 arm; do
  GOOS=linux GOARCH=$t CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "dist/diogows-linux-$t" .
done
# Windows: sem janela de console (roda em segundo plano, iniciado pela tarefa agendada do instalador)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -H=windowsgui" -o dist/diogows-windows-amd64.exe .
ls -lh dist
