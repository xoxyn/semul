package drafter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
)

type Prospect struct {
	Company           string
	Email             string
	IdentifiedProblem string
}

type DraftEmail struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func Draft(ctx context.Context, client *api.Client, prospect Prospect) (DraftEmail, error) {
	prompt := fmt.Sprintf(`
Write a concise, professional cold outreach email.

Our goal is to introduce our software and business workflow automation services,
explain how we could address the identified problem, and invite a conversation.

Company: %s
Identified problem: %s

Do not invent facts or claim solutions we cannot verify.
Return exactly:
Subject: <email subject>
Body:
<email body>
`, prospect.Company, prospect.IdentifiedProblem)

	req := &api.GenerateRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Format: json.RawMessage(`{
		"type": "object",
		"properties": {
			"subject": {"type": "string"},
			"body": {"type": "string"}
		},
		"required": ["subject", "body"]
	}`),
		Stream: new(bool),
	}

	var response strings.Builder

	err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response.WriteString(resp.Response)
		return nil
	})
	if err != nil {
		return DraftEmail{}, fmt.Errorf("generate email draft: %w", err)
	}

	fmt.Printf("RAW RESPONSE:\n%q\n", response.String())
	return parseDraft(response.String())
}

func parseDraft(response string) (DraftEmail, error) {
	var draft DraftEmail

	if err := json.Unmarshal([]byte(response), &draft); err != nil {
		return DraftEmail{}, fmt.Errorf("parse draft: %w", err)
	}

	if draft.Subject == "" || draft.Body == "" {
		return DraftEmail{}, fmt.Errorf("incomplete draft")
	}

	return draft, nil
}
