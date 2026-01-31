#!/bin/bash
set -e

echo "🚀 Iniciando build nativo para x86_64 (Intel)..."
echo "⚡ Isso deve ser rápido!"

# Executa o build mirando linux/amd64 (nativo da sua máquina)
docker build \
  -f Dockerfile.x86 \
  -t lambda-java-x86 \
  --output type=local,dest=. \
  .

echo "✅ Build concluído! O arquivo 'function.zip' (x86_64) está na pasta atual."
