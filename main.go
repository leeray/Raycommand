package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// Config struct for API keys and endpoints
type Config struct {
	DeepseekAPIKey string
	OpenAIApiKey   string
	DeepseekURL    string
	OpenAIURL      string
}

// LLMResponse is the structure to parse LLM API responses
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// API configuration
type APIConfig struct {
	Name   string
	URL    string
	APIKey string
}

var (
	deepseekAPI = APIConfig{
		Name:   "deepseek",
		URL:    "https://api.deepseek.com/chat/completions",
		APIKey: os.Getenv("DEEPSEEK_API_KEY"),
	}
	openaiAPI = APIConfig{
		Name:   "openai",
		URL:    "https://api.openai.com/v1/chat/completions",
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, using system environment variables.")
	}

	// Update API keys after loading .env
	deepseekAPI.APIKey = os.Getenv("DEEPSEEK_API_KEY")
	openaiAPI.APIKey = os.Getenv("OPENAI_API_KEY")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ray [your natural language command]")
		return
	}

	command := strings.Join(os.Args[1:], " ")
	apiConfig := deepseekAPI // Default to deepseek

	// Construct the prompt
	// Prepare system and directory information
	systemInfo, _ := exec.Command("uname", "-a").Output()
	currentDir, _ := os.Getwd()

	// Construct the prompt
	prompt := fmt.Sprintf("当前系统信息: %s, 当前目录: %s, 请帮我用bash命令完成以下任务:'%s'。1.仅输出代码。 2.不需要解释。", strings.TrimSpace(string(systemInfo)), currentDir, command)

	response, err := sendRequest(apiConfig, prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	// fmt.Printf("Response: %s\n", response)

	// Extract the command from the response
	extractedCommand, err := extractCommand(response)
	if err != nil {
		fmt.Printf("Failed to extract command: %v\n", err)
		return
	}

	// Output the extracted command
	// os.Stdout.WriteString(extractedCommand)

	// 执行命令
	err = executeCommand(extractedCommand)
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

type SystemInfo struct {
	OS  string
	Dir string
}

func getSystemInfo() (SystemInfo, error) {
	osInfo, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return SystemInfo{}, err
	}
	dir, err := os.Getwd()
	if err != nil {
		return SystemInfo{}, err
	}
	return SystemInfo{
		OS:  strings.TrimSpace(string(osInfo)),
		Dir: dir,
	}, nil
}

func sendRequest(apiConfig APIConfig, prompt string) (string, error) {
	if apiConfig.APIKey == "" {
		return "", errors.New("API key is missing. Please check your .env file or environment variables.")
	}

	// fmt.Printf("Calling LLM API: %s with key %s and prompt %q\n", apiConfig.URL, apiConfig.APIKey, prompt)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":             "deepseek-chat",
		"messages":          []map[string]string{{"role": "user", "content": prompt}},
		"frequency_penalty": 0,
		"max_tokens":        2048,
		"presence_penalty":  0,
		"response_format": map[string]string{
			"type": "text",
		},
		"stop":           nil,
		"stream":         false,
		"stream_options": nil,
		"temperature":    1,
		"top_p":          1,
		"tools":          nil,
		"tool_choice":    "none",
		"logprobs":       false,
		"top_logprobs":   nil,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiConfig.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiConfig.APIKey))

	// fmt.Printf("req: URL=%s, Header=%#v, Body=%s\n", req.URL, req.Header, string(requestBody))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", errors.New(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractCommand(response string) (string, error) {
	// Parse the JSON response
	var parsedResponse map[string]interface{}
	err := json.Unmarshal([]byte(response), &parsedResponse)
	if err != nil {
		return "", err
	}

	// Navigate to the choices and extract the command
	choices, ok := parsedResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("no choices found in response")
	}

	firstChoice := choices[0].(map[string]interface{})
	content, ok := firstChoice["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return "", errors.New("failed to extract content from response")
	}

	// Extract the command from the response content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !strings.HasPrefix(line, "```bash") {
			return line, nil // Return the first non-comment line
		}
	}

	return "", errors.New("no valid command found in response")
}

func executeCommand(command string) error {
	fmt.Println("\n正在执行以下命令...")
	fmt.Println(command)
	fmt.Println("\n")
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
