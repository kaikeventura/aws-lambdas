#!/bin/bash
set -e

echo "🐍 Iniciando build da Lambda Python..."

# Limpa builds anteriores
rm -rf package lambda.zip

# Cria diretório de pacote
mkdir package

# Instala dependências no diretório local
echo "📦 Instalando dependências..."
pip install -r requirements.txt -t package/

# Copia o código fonte
echo "📄 Copiando código fonte..."
cp lambda_function.py package/

# Cria o zip
echo "🤐 Zipando..."
cd package
zip -r ../lambda.zip .
cd ..

echo "✅ Build concluído! Arquivo 'lambda.zip' gerado."
