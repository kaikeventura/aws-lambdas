#!/bin/bash
set -e

echo "🚀 Iniciando build nativo para ARM64 (AWS Graviton)..."
echo "⚠️  Atenção: Como você está em uma máquina Intel, isso usará emulação QEMU."
echo "☕ Pegue um café, isso vai demorar bastante (10-30 min)..."

# Habilita emulação QEMU para Docker (caso não esteja habilitado)
if ! docker buildx inspect | grep -q "linux/arm64"; then
    echo "🔧 Configurando QEMU para suporte multi-arquitetura..."
    docker run --privileged --rm tonistiigi/binfmt --install all
fi

# Cria um builder se não existir
if ! docker buildx inspect mybuilder > /dev/null 2>&1; then
    echo "🏗️  Criando builder 'mybuilder'..."
    docker buildx create --name mybuilder --use
fi

# Executa o build mirando linux/arm64
# --load não funciona com multi-platform em alguns drivers, então usamos --output type=local
echo "🔨 Compilando..."
docker buildx build \
  --platform linux/arm64 \
  -f Dockerfile.build \
  -t lambda-java-arm64 \
  --output type=local,dest=. \
  .

echo "✅ Build concluído! O arquivo 'function.zip' (ARM64) está na pasta atual."
