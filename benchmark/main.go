package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
	"sync"
)

type EndpointConfig struct {
	URL    string
	APIKey string
	Body   string
}

type APITestConfig struct {
	Name           string
	SignupEndpoint EndpointConfig
	SigninEndpoint EndpointConfig
	AuthEndpoint   EndpointConfig
}

type SigninResponse struct {
	Token string `json:"token"`
}

var (
	javaBaseURL = "JAVA_API_GATEWAY_URL"
	javaAPIKey  = "JAVA_API_KEY"

	pythonBaseURL = "PYTHON_API_GATEWAY_URL"
	pythonAPIKey  = "PYTHON_API_KEY"
)

var javaAPI = APITestConfig{
	Name: "Java",
	SignupEndpoint: EndpointConfig{
		URL:    javaBaseURL + "/api/signup",
		APIKey: javaAPIKey,
		Body:   `{"email": "%s", "password": "senha", "name": "Java User"}`,
	},
	SigninEndpoint: EndpointConfig{
		URL:    javaBaseURL + "/api/signin",
		APIKey: javaAPIKey,
		Body:   `{"email": "%s", "password": "senha"}`,
	},
	AuthEndpoint: EndpointConfig{
		URL:    javaBaseURL + "/api/authentication",
		APIKey: javaAPIKey,
	},
}

var pythonAPI = APITestConfig{
	Name: "Python",
	SignupEndpoint: EndpointConfig{
		URL:    pythonBaseURL + "/signup",
		APIKey: pythonAPIKey,
		Body:   `{"email": "%s", "password": "MinhaSenhaForte123!", "name": "Kaike Teste"}`,
	},
	SigninEndpoint: EndpointConfig{
		URL:    pythonBaseURL + "/signin",
		APIKey: pythonAPIKey,
		Body:   `{"email": "%s", "password": "MinhaSenhaForte123!"}`,
	},
	AuthEndpoint: EndpointConfig{
		URL:    pythonBaseURL + "/authentication",
		APIKey: pythonAPIKey,
	},
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Iniciando Benchmark para API Java ===")
	runFullBenchmark(javaAPI)

	fmt.Println("\n=== Iniciando Benchmark para API Python ===")
	runFullBenchmark(pythonAPI)
}

func runFullBenchmark(api APITestConfig) {
	fmt.Printf("\n--- Testando fluxo completo para: %s ---\n", api.Name)

	// Cenário 1: Pico de requisições
	fmt.Println("Cenário 1: Pico de requisições (50 flows/s por 20s)")
	runFlowLoadTest(api, 20*time.Second, 50)

	// Pausa
	fmt.Println("Pausando por 10 segundos...")
	time.Sleep(10 * time.Second)

	// Cenário 2: Nova carga de requisições
	fmt.Println("Cenário 2: Carga de requisições (25 flows/s por 17s)")
	runFlowLoadTest(api, 17*time.Second, 25)

	// Pausa de 5 minutos
	fmt.Println("Pausando por 5 minutos...")
	time.Sleep(5 * time.Minute)

	// Cenário 3: Outra carga de requisições
	fmt.Println("Cenário 3: Carga de requisições (40 flows/s por 6s)")
	runFlowLoadTest(api, 6*time.Second, 40)

	fmt.Printf("--- Benchmark finalizado para: %s ---\n", api.Name)
}

func runFlowLoadTest(api APITestConfig, duration time.Duration, flowsPerSecond int) {
	var wg sync.WaitGroup
	requests := make(chan struct{})
	client := &http.Client{Timeout: 20 * time.Second} // Timeout para o fluxo completo

	for i := 0; i < 50; i++ { // 50 workers concorrentes
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range requests {
				executeFullFlow(client, api)
			}
		}()
	}

	ticker := time.NewTicker(time.Second / time.Duration(flowsPerSecond))
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			close(requests)
			wg.Wait()
			return
		case <-ticker.C:
			requests <- struct{}{}
		}
	}
}

func executeFullFlow(client *http.Client, api APITestConfig) {
	// 1. Signup
	email := fmt.Sprintf("user_%s@test.com", randomString(12))
	signupBody := fmt.Sprintf(api.SignupEndpoint.Body, email)

	err := doPost(client, api.SignupEndpoint.URL, api.SignupEndpoint.APIKey, "", signupBody)
	if err != nil {
		return
	}

	// 2. Signin
	signinBody := fmt.Sprintf(api.SigninEndpoint.Body, email)
	respBody, err := doPostAndReadBody(client, api.SigninEndpoint.URL, api.SigninEndpoint.APIKey, "", signinBody)
	if err != nil {
		return
	}

	var signinResp SigninResponse
	if err := json.Unmarshal(respBody, &signinResp); err != nil {
		return
	}
	token := signinResp.Token
	if token == "" {
		return
	}

	// 3. Authentication
	err = doPost(client, api.AuthEndpoint.URL, api.AuthEndpoint.APIKey, "Bearer "+token, "")
	if err != nil {
		return
	}
}

func doPost(client *http.Client, url, apiKey, authToken, body string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

func doPostAndReadBody(client *http.Client, url, apiKey, authToken, body string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}
	return respBody, nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
