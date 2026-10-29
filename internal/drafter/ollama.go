
package drafter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/xoxyn/semul/internal/domain"
)

func Draft(
	ctx context.Context,
	client *api.Client,
	sender domain.Sender,
	prospect domain.Prospect,
) (domain.DraftEmail, error) {
	prompt := fmt.Sprintf(`
Write a concise, professional cold outreach email.

SENDER:
Name: %s
Company: %s

RECIPIENT:
Company: %s
Email: %s
Identified problem: %s

OBJECTIVE:
Introduce the sender's software and business workflow automation services.
Explain how these services could address the identified problem.
Invite the recipient to a brief conversation.

INSTRUCTIONS:
- Address the recipient using their company name, not a placeholder.
- Sign off using the sender's actual name and company.
- Never use placeholders such as [Your Name] or [Recipient].
- Never invent a recipient's personal name.
- Never confuse the sender's company with the recipient's company.
- Do not invent facts, experience, or results.
- Keep the email concise and professional.

Return a JSON object containing "subject" and "body".
`, sender.Name, sender.Company,
		prospect.Company, prospect.Email, prospect.IdentifiedProblem)

	req := &api.GenerateRequest{
		Model:  "llama3.2",
		Prompt: strings.TrimSpace(prompt),
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
		return domain.DraftEmail{}, fmt.Errorf("generate email draft: %w", err)
	}

	return parseDraft(response.String())
}

func parseDraft(response string) (domain.DraftEmail, error) {
	var draft domain.DraftEmail

	if err := json.Unmarshal([]byte(response), &draft); err != nil {
		return domain.DraftEmail{}, fmt.Errorf("parse draft: %w", err)
	}

	if strings.TrimSpace(draft.Subject) == "" ||
		strings.TrimSpace(draft.Body) == "" {
		return domain.DraftEmail{}, fmt.Errorf("incomplete draft")
	}

	return draft, nil
}
