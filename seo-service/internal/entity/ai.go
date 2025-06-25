package entity

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiSEORequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type AiSEOResponse struct {
	Reply    string `json:"reply"`
	Model    string `json:"model"`
	Created  int64  `json:"created"`
	Response string `json:"response"`
}
