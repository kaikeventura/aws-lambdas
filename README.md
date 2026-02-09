# AWS Lambdas Benchmark Project

Este repositório contém um projeto completo para comparar o desempenho de AWS Lambdas implementadas em **Java (Quarkus)**, **Python**, **Go** e **Rust**. O projeto inclui a infraestrutura como código (Terraform), o código fonte das funções e uma ferramenta de benchmark.

## Estrutura do Projeto

A estrutura de diretórios é organizada da seguinte forma:

```
/
├── benchmark/      # Ferramenta de teste de carga escrita em Go
├── go/             # Implementação da Lambda em Go
├── infra/          # Infraestrutura como código (Terraform)
├── java/           # Implementação da Lambda em Java (Quarkus)
├── python/         # Implementação da Lambda em Python
├── rust/           # Implementação da Lambda em Rust
└── README.md       # Documentação do projeto
```

---

## 1. Java (Quarkus)

A implementação Java utiliza o framework **Quarkus** para otimizar o tempo de inicialização e consumo de memória, ideal para ambientes Serverless.

*   **Localização:** `/java`
*   **Framework:** Quarkus 3.15.1
*   **Java Version:** 21
*   **Build Tool:** Maven

---

## 2. Python

A implementação Python é uma função Lambda padrão, utilizando bibliotecas leves para manter o cold start baixo.

*   **Localização:** `/python`
*   **Runtime:** Python 3.12
*   **Principais Dependências (`requirements.txt`):**
    *   `boto3`: SDK da AWS para interagir com DynamoDB.
    *   `PyJWT`: Manipulação de tokens JWT.
    *   `bcrypt`: Hashing de senhas.

---

## 3. Go

A implementação em Go é compilada para um binário nativo e executada no runtime `provided.al2`.

*   **Localização:** `/go`
*   **Runtime:** `provided.al2`
*   **Build:** O binário é compilado estaticamente com `CGO_ENABLED=0` para evitar problemas de compatibilidade com a `glibc`.
*   **Principais Dependências:**
    *   `github.com/aws/aws-lambda-go`: Runtime da AWS Lambda.
    *   `github.com/aws/aws-sdk-go`: SDK da AWS.
    *   `github.com/golang-jwt/jwt/v5`: Manipulação de JWT.
    *   `golang.org/x/crypto/bcrypt`: Hashing de senhas.

---

## 4. Rust

A implementação em Rust também é compilada para um binário nativo estático, visando máxima performance e segurança.

*   **Localização:** `/rust`
*   **Runtime:** `provided.al2`
*   **Build:** O binário é compilado para o target `x86_64-unknown-linux-musl` para garantir compatibilidade com o ambiente Lambda. O build é feito via Docker para não depender de ferramentas no host.
*   **Principais Dependências:**
    *   `lambda_runtime`: Runtime da AWS Lambda.
    *   `aws-sdk-dynamodb`: SDK da AWS.
    *   `jsonwebtoken`: Manipulação de JWT.
    *   `bcrypt`: Hashing de senhas.
    *   `tokio`: Runtime assíncrono.

---

## 5. Infraestrutura (Terraform)

A infraestrutura é provisionada utilizando **Terraform**, garantindo reprodutibilidade e gerenciamento de estado.

*   **Localização:** `/infra/terraform`
*   **Provider:** AWS (~> 5.0)
*   **Região:** `us-east-1`

### Módulos
*   `modules/dynamodb`: Criação das tabelas DynamoDB.
*   `modules/lambda-java`: Provisionamento da Lambda Java.
*   `modules/lambda-python`: Provisionamento da Lambda Python.
*   `modules/lambda-go`: Provisionamento da Lambda Go.
*   `modules/lambda-rust`: Provisionamento da Lambda Rust.
*   `modules/api-gateway`: Configuração do API Gateway para expor as Lambdas.

### Ambientes
*   `envs/dev`: Configuração específica para o ambiente de desenvolvimento.

---

## 6. Benchmark (Go)

Uma ferramenta personalizada escrita em **Go** para executar testes de carga e comparar o desempenho das implementações.

*   **Localização:** `/benchmark`
*   **Linguagem:** Go
*   **Arquivo Principal:** `main.go`

### Como Executar
1.  Certifique-se de ter Go instalado.
2.  Navegue até a pasta `benchmark`.
3.  Configure as variáveis de ambiente ou edite as constantes de URL e API Key no arquivo `main.go` com os valores do seu ambiente implantado (obtidos dos outputs do Terraform).
4.  Execute:
    ```bash
    go run main.go
    ```

---

## 7. Resultados do Benchmark (CloudWatch)

**Nota:** Os resultados a seguir comparam apenas as implementações em **Java** e **Python**. Testes para Go e Rust serão adicionados futuramente.

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

#### 2. Python: Latency & Integration Latency
Latência total e de integração para a stack Python.

| Metric | Min (ms) | Max (ms) | Avg (ms) | P99 (ms) |
| :--- | :--- | :--- | :--- | :--- |
| **Integration Latency** | 10 | 3808 | 1399.07 | 3538.98 |
| **Total Latency** | 14 | 3813 | 1403.72 | 3538.98 |

![API Gateway Python Latency.png](prints/API%20Gateway%20Python%20Latency.png)

---

## 8. Conclusão e Comparativo Final (Java vs Python)

Com base nos dados coletados, podemos fazer uma análise detalhada entre as duas implementações:

### Desempenho (Latência e Throughput)
*   **Java (Quarkus)** demonstrou um desempenho significativamente superior.
    *   **Duração Média:** ~471ms vs ~1350ms do Python. O Java foi quase **3x mais rápido** em média.
    *   **Cold Start:** O Java (com Quarkus) teve um cold start médio de ~524ms, enquanto o Python teve ~1017ms. Isso é surpreendente, pois geralmente Python tem cold starts menores, mas mostra a eficiência do Quarkus.
    *   **Throughput:** O Java processou mais que o dobro de invocações (3171 vs 1479) no mesmo período de teste, indicando maior capacidade de vazão.

### Consumo de Recursos
*   **Memória:** O Java consumiu mais memória (Max ~130MB) comparado ao Python (~88MB). Isso é esperado dada a JVM, mas o valor é baixo para Java, validando o uso do Quarkus.

### Custo
*   **Custo Total:** O custo estimado para o Java foi **menor** ($0.0061 vs $0.0081), mesmo processando **mais que o dobro** de requisições.
*   **Eficiência de Custo:** Como o custo da Lambda é baseado em (Duração * Memória), a rapidez do Java compensou o uso ligeiramente maior de memória.

### Veredito: Qual a melhor escolha?

**Vencedor (Java vs Python): Java (Quarkus)**

Para este cenário de API REST com DynamoDB, a implementação em **Java com Quarkus** é a escolha superior.

*   ✅ **Mais Rápido:** Respostas 3x mais rápidas para o usuário final.
*   ✅ **Mais Barato:** Menor custo por transação devido à menor duração de execução.
*   ✅ **Escalabilidade:** Demonstrou capacidade de lidar com maior volume de carga (throughput).
