# AWS Lambdas Benchmark Project

Este repositório contém um projeto completo para comparar o desempenho de AWS Lambdas implementadas em **Java (Quarkus)** e **Python**. O projeto inclui a infraestrutura como código (Terraform), o código fonte das funções e uma ferramenta de benchmark em Go.

## Estrutura do Projeto

A estrutura de diretórios é organizada da seguinte forma:

```
/
├── benchmark/      # Ferramenta de teste de carga escrita em Go
├── infra/          # Infraestrutura como código (Terraform)
├── java/           # Implementação da Lambda em Java (Quarkus)
├── python/         # Implementação da Lambda em Python
└── README.md       # Documentação do projeto
```

---

## 1. Java (Quarkus)

A implementação Java utiliza o framework **Quarkus** para otimizar o tempo de inicialização e consumo de memória, ideal para ambientes Serverless.

*   **Localização:** `/java`
*   **Framework:** Quarkus 3.15.1
*   **Java Version:** 21
*   **Build Tool:** Maven
*   **Principais Dependências:**
    *   `quarkus-amazon-lambda-rest`: Integração com AWS Lambda e API Gateway.
    *   `quarkus-resteasy-jackson`: Serialização JSON.
    *   `quarkus-amazon-dynamodb-enhanced`: Cliente DynamoDB otimizado.
    *   `quarkus-smallrye-jwt`: Autenticação JWT.
    *   `jbcrypt`: Hashing de senhas.
    *   `lombok`: Redução de boilerplate code.

### Build e Deploy
O projeto suporta build nativo e JVM. Scripts auxiliares estão disponíveis:
*   `build-x86.sh`: Compila para arquitetura x86.
*   `build-arm64.sh`: Compila para arquitetura ARM64 (Graviton).

---

## 2. Python

A implementação Python é uma função Lambda padrão, utilizando bibliotecas leves para manter o cold start baixo.

*   **Localização:** `/python`
*   **Runtime:** Python 3.x
*   **Principais Dependências (`requirements.txt`):**
    *   `boto3`: SDK da AWS para interagir com DynamoDB.
    *   `PyJWT`: Manipulação de tokens JWT.
    *   `bcrypt`: Hashing de senhas.

### Estrutura
*   `lambda_function.py`: Ponto de entrada da função.
*   `build.sh`: Script para empacotar a função e dependências em um arquivo ZIP.

---

## 3. Infraestrutura (Terraform)

A infraestrutura é provisionada utilizando **Terraform**, garantindo reprodutibilidade e gerenciamento de estado.

*   **Localização:** `/infra/terraform`
*   **Provider:** AWS (~> 5.0)
*   **Região:** `us-east-1`

### Módulos
*   `modules/dynamodb`: Criação das tabelas DynamoDB (ex: tabela `users`).
*   `modules/lambda-java`: Provisionamento da Lambda Java.
*   `modules/lambda-python`: Provisionamento da Lambda Python.
*   `modules/api-gateway`: Configuração do API Gateway para expor as Lambdas.

### Ambientes
*   `envs/dev`: Configuração específica para o ambiente de desenvolvimento.

---

## 4. Benchmark (Go)

Uma ferramenta personalizada escrita em **Go** para executar testes de carga e comparar o desempenho das duas implementações.

*   **Localização:** `/benchmark`
*   **Linguagem:** Go
*   **Arquivo Principal:** `main.go`

### Funcionalidades
O benchmark executa um fluxo completo de autenticação (Signup -> Signin -> Auth) simulando usuários reais.

### Cenários de Teste Atuais
O teste é dividido em três fases para simular diferentes cargas:
1.  **Cenário 1:** 50 flows/s por 20 segundos.
2.  **Pausa:** 10 segundos.
3.  **Cenário 2:** 25 flows/s por 17 segundos.
4.  **Pausa:** 5 minutos.
5.  **Cenário 3:** 40 flows/s por 6 segundos.

**Workers:** 50 goroutines concorrentes são utilizadas para gerar a carga.

### Como Executar
1.  Certifique-se de ter Go instalado.
2.  Navegue até a pasta `benchmark`.
3.  Configure as variáveis de ambiente ou edite as constantes `javaBaseURL`, `javaAPIKey`, `pythonBaseURL`, `pythonAPIKey` no arquivo `main.go` com os valores do seu ambiente implantado.
4.  Execute:
    ```bash
    go run main.go
    ```

---

## 5. Resultados do Benchmark (CloudWatch)

Abaixo estão os resultados coletados via CloudWatch Dashboard para uma execução de teste.

### Lambda Metrics

#### 1. Statistics (Invocations, Duration, Cold Start, Memory)
Comparativo geral de invocações, duração média/máxima, cold starts e uso de memória.

| Metric | Python | Java |
| :--- | :--- | :--- |
| **Total Invocations** | 1479 | 3171 |
| **Avg Duration (ms)** | 1350.55 | 471.67 |
| **Max Duration (ms)** | 2413.19 | 1696.10 |
| **Avg Cold Start (ms)** | 1017.76 | 524.38 |
| **Max Cold Start (ms)** | 1151.60 | 679.99 |
| **Cold Start Count** | 50 | 75 |
| **Max Mem Used (MB)** | 87.74 | 129.70 |
| **Allocated Mem (MB)** | 244.14 | 244.14 |

![Lambda Statistics.png](prints/Lambda%20Statistics.png)

#### 2. Invocations & Errors
Série temporal de invocações e erros para ambas as funções.

| Timestamp | Java Invocations | Java Errors | Python Invocations | Python Errors |
| :--- | :--- | :--- | :--- | :--- |
| 2026/02/07 13:04:00 | 3171 | 0 | 1479 | 0 |

![Lambda Invocations Errors.png](prints/Lambda%20Invocations%20Errors.png)

#### 3. Duration & P99
Comparativo de duração média e percentil 99 (P99).

| Timestamp | Java Avg Duration (ms) | Java P99 (ms) | Python Avg Duration (ms) | Python P99 (ms) |
| :--- | :--- | :--- | :--- | :--- |
| 2026/02/07 13:04:00 | 484.08 | 1996.28 | 1350.55 | 2255.40 |

![Lambda Duration P99.png](prints/Lambda%20Duration%20P99.png)

#### 4. Throttles & Concurrent Executions
Monitoramento de throttling e execuções concorrentes.

| Timestamp | Java Throttles | Java Concurrent | Python Throttles | Python Concurrent |
| :--- | :--- | :--- | :--- | :--- |
| 2026/02/07 13:04:00 | 0 | 110386 | 0 | 55354 |

![Lambda Throttles Concurrent.png](prints/Lambda%20Throttles%20Concurrent.png)

#### 5. Estimated Cost (USD)
Custo estimado baseado na duração cobrada e memória alocada.

| Function | Cost (USD) |
| :--- | :--- |
| Python | $0.008142 |
| Java | $0.006106 |

![Lambda Cost.png](prints/Lambda%20Cost.png)

### API Gateway Metrics

#### 1. Java: Count, 5XX & 4XX Errors
Volume de requisições e erros no API Gateway para a stack Java.

| Timestamp | Count | 5XX Error | 4XX Error |
| :--- | :--- | :--- | :--- |
| 2026/02/07 13:02:00 | 3171 | 0 | 0 |

![API Gateway Java Errors.png](prints/API%20Gateway%20Java%20Errors.png)

#### 2. Python: Count, 5XX & 4XX Errors
Volume de requisições e erros no API Gateway para a stack Python.

| Timestamp | Count | 5XX Error | 4XX Error |
| :--- | :--- | :--- | :--- |
| 2026/02/07 13:03:00 | 1479 | 0 | 0 |

![API Gateway Python Errors.png](prints/API%20Gateway%20Python%20Errors.png)

#### 3. Java: Latency & Integration Latency
Latência total e de integração para a stack Java.

| Metric | Min (ms) | Max (ms) | Avg (ms) | P99 (ms) |
| :--- | :--- | :--- | :--- | :--- |
| **Integration Latency** | 10 | 2653 | 503.35 | 2280.50 |
| **Total Latency** | 15 | 2694 | 507.85 | 2295.95 |

![API Gateway Java Latency.png](prints/API%20Gateway%20Java%20Latency.png)

#### 4. Python: Latency & Integration Latency
Latência total e de integração para a stack Python.

| Metric | Min (ms) | Max (ms) | Avg (ms) | P99 (ms) |
| :--- | :--- | :--- | :--- | :--- |
| **Integration Latency** | 10 | 3808 | 1399.07 | 3538.98 |
| **Total Latency** | 14 | 3813 | 1403.72 | 3538.98 |

![API Gateway Python Latency.png](prints/API%20Gateway%20Python%20Latency.png)

---

## 6. Conclusão e Comparativo Final

Com base nos dados coletados, podemos fazer uma análise detalhada entre as duas implementações:

### Desempenho (Latência e Throughput)
*   **Java (Quarkus)** demonstrou um desempenho significativamente superior.
    *   **Duração Média:** ~471ms vs ~1350ms do Python. O Java foi quase **3x mais rápido** em média.
    *   **Cold Start:** O Java (com Quarkus) teve um cold start médio de ~524ms, enquanto o Python teve ~1017ms. Isso é surpreendente, pois geralmente Python tem cold starts menores, mas mostra a eficiência do Quarkus.
    *   **Throughput:** O Java processou mais que o dobro de invocações (3171 vs 1479) no mesmo período de teste, indicando maior capacidade de vazão.

### Consumo de Recursos
*   **Memória:** O Java consumiu mais memória (Max ~130MB) comparado ao Python (~88MB). Isso é esperado dada a JVM, mas o valor é baixo para Java, validando o uso do Quarkus.
*   **Concorrência:** O Java atingiu picos de concorrência muito maiores, o que é natural dado o maior volume de requisições processadas.

### Custo
*   **Custo Total:** O custo estimado para o Java foi **menor** ($0.0061 vs $0.0081), mesmo processando **mais que o dobro** de requisições.
*   **Eficiência de Custo:** Como o custo da Lambda é baseado em (Duração * Memória), a rapidez do Java compensou o uso ligeiramente maior de memória. O Python, sendo mais lento, acabou custando mais por transação.

### Veredito: Qual a melhor escolha?

**Vencedor: Java (Quarkus)**

Para este cenário de API REST com DynamoDB, a implementação em **Java com Quarkus** é a escolha superior.

*   ✅ **Mais Rápido:** Respostas 3x mais rápidas para o usuário final.
*   ✅ **Mais Barato:** Menor custo por transação devido à menor duração de execução.
*   ✅ **Escalabilidade:** Demonstrou capacidade de lidar com maior volume de carga (throughput).

A implementação em Python, embora mais simples e com menor consumo de memória, sofreu com tempos de execução mais longos, o que impactou diretamente a latência e o custo final.
