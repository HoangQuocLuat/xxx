package groq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"ai-monitor/tracer"
)

const (
	groqURL = "https://api.groq.com/openai/v1/chat/completions"
	model   = "llama-3.3-70b-versatile"
)

type AnalyzeRequest struct {
	ErrorMessage  string
	Exception     string
	URL           string
	MainSnippet   tracer.CodeSnippet
	TraceSnippets []tracer.CodeSnippet
}

type Solution struct {
	RootCause string `json:"root_cause"`
	File      string `json:"file"`
	Line      string `json:"line"`
	Solution  string `json:"solution"`
	CodeFix   string `json:"code_fix"`
}

type groqRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func Analyze(req AnalyzeRequest) (*Solution, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY not set")
	}

	prompt := buildPrompt(req)

	body, _ := json.Marshal(groqRequest{
		Model:       model,
		Temperature: 0.1,
		Messages: []message{
			{
				Role: "system",
				Content: `You are a senior PHP developer analyzing error logs.
Respond ONLY with a valid JSON object, no markdown, no explanation outside JSON:
{
  "root_cause": "brief explanation of why the error happened",
  "file": "the file where the error originates",
  "line": "line number",
  "solution": "step by step fix instructions",
  "code_fix": "the actual SQL or PHP code to fix it, empty string if not applicable"
}`,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	httpReq, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var groqResp groqResponse
	if err := json.Unmarshal(raw, &groqResp); err != nil {
		return nil, fmt.Errorf("parse groq response: %w", err)
	}
	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from groq")
	}

	content := groqResp.Choices[0].Message.Content

	// Strip markdown code fences nếu model vẫn wrap
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var solution Solution
	if err := json.Unmarshal([]byte(content), &solution); err != nil {
		// Fallback: trả raw text vào RootCause để không mất thông tin
		return &Solution{RootCause: content}, nil
	}

	return &solution, nil
}

func buildPrompt(req AnalyzeRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error: %s\n", req.ErrorMessage))
	sb.WriteString(fmt.Sprintf("Exception class: %s\n", req.Exception))
	sb.WriteString(fmt.Sprintf("Request: %s\n\n", req.URL))

	sb.WriteString("=== Error location ===\n")
	sb.WriteString(fmt.Sprintf("File: %s (line %d)\n", req.MainSnippet.File, req.MainSnippet.ErrLine))
	sb.WriteString(strings.Join(req.MainSnippet.Lines, "\n"))
	sb.WriteString("\n")

	if len(req.TraceSnippets) > 0 {
		sb.WriteString("\n=== Stack trace snippets ===\n")
		for _, s := range req.TraceSnippets {
			if len(s.Lines) == 0 {
				continue
			}
			sb.WriteString(fmt.Sprintf("\nFile: %s (line %d)\n", s.File, s.ErrLine))
			sb.WriteString(strings.Join(s.Lines, "\n"))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
