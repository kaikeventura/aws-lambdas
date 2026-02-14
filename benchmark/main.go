package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
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

	goBaseURL = "GO_API_GATEWAY_URL"
	goAPIKey  = "GO_API_KEY"

	rustBaseURL = "RUST_API_GATEWAY_URL"
	rustAPIKey  = "RUST_API_KEY"

	dotnetBaseURL = "DOTNET_API_GATEWAY_URL"
	dotnetAPIKey  = "DOTNET_API_KEY"

	nodejsBaseURL = "NODEJS_API_GATEWAY_URL"
	nodejsAPIKey  = "NODEJS_API_KEY"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	apisToTest := []APITestConfig{
		{
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
		},
		{
			Name: "Python",
			SignupEndpoint: EndpointConfig{
				URL:    pythonBaseURL + "/signup",
				APIKey: pythonAPIKey,
				Body:   `{"email": "%s", "password": "MinhaSenhaForte123!", "name": "Python User"}`,
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
		},
		{
			Name: "Go",
			SignupEndpoint: EndpointConfig{
				URL:    goBaseURL + "/signup",
				APIKey: goAPIKey,
				Body:   `{"email": "%s", "password": "password123", "name": "Go User"}`,
			},
			SigninEndpoint: EndpointConfig{
				URL:    goBaseURL + "/signin",
				APIKey: goAPIKey,
				Body:   `{"email": "%s", "password": "password123"}`,
			},
			AuthEndpoint: EndpointConfig{
				URL:    goBaseURL + "/authentication",
				APIKey: goAPIKey,
			},
		},
		{
			Name: "Rust",
			SignupEndpoint: EndpointConfig{
				URL:    rustBaseURL + "/signup",
				APIKey: rustAPIKey,
				Body:   `{"email": "%s", "password": "password123", "name": "Rust User"}`,
			},
			SigninEndpoint: EndpointConfig{
				URL:    rustBaseURL + "/signin",
				APIKey: rustAPIKey,
				Body:   `{"email": "%s", "password": "password123"}`,
			},
			AuthEndpoint: EndpointConfig{
				URL:    rustBaseURL + "/authentication",
				APIKey: rustAPIKey,
			},
		},
		{
			Name: "DotNet",
			SignupEndpoint: EndpointConfig{
				URL:    dotnetBaseURL + "/signup",
				APIKey: dotnetAPIKey,
				Body:   `{"email": "%s", "password": "password123", "name": "DotNet User"}`,
			},
			SigninEndpoint: EndpointConfig{
				URL:    dotnetBaseURL + "/signin",
				APIKey: dotnetAPIKey,
				Body:   `{"email": "%s", "password": "password123"}`,
			},
			AuthEndpoint: EndpointConfig{
				URL:    dotnetBaseURL + "/authentication",
				APIKey: dotnetAPIKey,
			},
		},
		{
			Name: "NodeJS",
			SignupEndpoint: EndpointConfig{
				URL:    nodejsBaseURL + "/signup",
				APIKey: nodejsAPIKey,
				Body:   `{"email": "%s", "password": "password123", "name": "NodeJS User"}`,
			},
			SigninEndpoint: EndpointConfig{
				URL:    nodejsBaseURL + "/signin",
				APIKey: nodejsAPIKey,
				Body:   `{"email": "%s", "password": "password123"}`,
			},
			AuthEndpoint: EndpointConfig{
				URL:    nodejsBaseURL + "/authentication",
				APIKey: nodejsAPIKey,
			},
		},
	}

	for _, api := range apisToTest {
		fmt.Printf("\n=== Iniciando Benchmark para API %s ===\n", api.Name)
		runFullBenchmark(api)
	}
}

func runFullBenchmark(api APITestConfig) {
	fmt.Printf("\n--- Testando fluxo completo para: %s ---\n", api.Name)

	fmt.Println("Cenário 1: Pico de requisições (50 flows/s por 20s)")
	runFlowLoadTest(api, 20*time.Second, 50)

	fmt.Println("Pausando por 10 segundos...")
	time.Sleep(10 * time.Second)

	fmt.Println("Cenário 2: Carga de requisições (25 flows/s por 17s)")
	runFlowLoadTest(api, 17*time.Second, 25)

	fmt.Println("Pausando por 5 minutos...")
	time.Sleep(5 * time.Minute)

	fmt.Println("Cenário 3: Carga de requisições (40 flows/s por 6s)")
	runFlowLoadTest(api, 6*time.Second, 40)

	fmt.Printf("--- Benchmark finalizado para: %s ---\n", api.Name)
}

func runFlowLoadTest(api APITestConfig, duration time.Duration, flowsPerSecond int) {
	var wg sync.WaitGroup
	requests := make(chan struct{})
	client := &http.Client{Timeout: 20 * time.Second}

	// Number of concurrent workers
	const numWorkers = 50
	for i := 0; i < numWorkers; i++ {
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
	email := fmt.Sprintf("user_%s@test.com", randomString(12))
	signupBody := fmt.Sprintf(api.SignupEndpoint.Body, email)

	err := doPost(client, api.SignupEndpoint.URL, api.SignupEndpoint.APIKey, "", signupBody)
	if err != nil {
		// Log error but continue, as we are in a load test
		// fmt.Printf("Error during signup for %s: %v\n", api.Name, err)
		return
	}

	signinBody := fmt.Sprintf(api.SigninEndpoint.Body, email)
	respBody, err := doPostAndReadBody(client, api.SigninEndpoint.URL, api.SigninEndpoint.APIKey, "", signinBody)
	if err != nil {
		// fmt.Printf("Error during signin for %s: %v\n", api.Name, err)
		return
	}

	var signinResp SigninResponse
	if err := json.Unmarshal(respBody, &signinResp); err != nil {
		// fmt.Printf("Error unmarshalling signin response for %s: %v\n", api.Name, err)
		return
	}
	token := signinResp.Token
	if token == "" {
		// fmt.Printf("Token not found in signin response for %s\n", api.Name)
		return
	}

	err = doPost(client, api.AuthEndpoint.URL, api.AuthEndpoint.APIKey, "Bearer "+token, "")
	if err != nil {
		// fmt.Printf("Error during authentication for %s: %v\n", api.Name, err)
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
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
