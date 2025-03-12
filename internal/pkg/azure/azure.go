package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestTool struct {
	Type     string `json:"type"`
	Function any    `json:"function"`
}

type RequestResponseFormat struct {
	Type       string `json:"type"`
	JsonSchema any    `json:"json_schema"`
}

type RequestBody struct {
	Messages       []Message             `json:"messages"`
	Tools          []RequestTool         `json:"tools"`
	ResponseFormat RequestResponseFormat `json:"response_format"`
}

type ResponseTool struct {
	Function struct {
		Arguments string `json:"arguments"`
		Name      string `json:"name"`
	} `json:"function"`
	Id   string `json:"id"`
	Type string `json:"type"`
}

type ResponseChoice struct {
	Message struct {
		Content   string         `json:"content"`
		Role      string         `json:"role"`
		ToolCalls []ResponseTool `json:"tool_calls"`
	} `json:"message"`
}

type APIResponse struct {
	Choices []ResponseChoice `json:"choices"`
}

type ChatAPI struct {
	apiUrl         string
	token          string
	tools          []RequestTool
	responseFormat RequestResponseFormat
}

// NewAzureChatAPI creates a chat client that utilizes Azure OpenAPI service
func NewAzureChatAPI(apiUrl string, token string, responseFormat RequestResponseFormat, tools []RequestTool) *ChatAPI {
	return &ChatAPI{
		apiUrl:         apiUrl,
		token:          token,
		tools:          tools,
		responseFormat: responseFormat,
	}
}

// Chat with Azure Chat API
func (a *ChatAPI) Chat(messages []Message) (*APIResponse, error) {
	requestBody := RequestBody{Messages: messages,
		Tools:          a.tools,
		ResponseFormat: a.responseFormat,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", a.apiUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", a.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return &apiResponse, nil
}
