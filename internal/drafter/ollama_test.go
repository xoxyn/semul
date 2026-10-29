package drafter

import (
	"context"
	"fmt"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestParseDraft(t *testing.T) {
	response := `{
        "subject": "Streamline Logistics Reporting with Automation",
        "body": "Dear Recipient,\n\nWe help businesses automate manual workflows."
    }`

	draft, err := parseDraft(response)
	if err != nil {
		t.Fatal(err)
	}

	if draft.Subject != "Streamline Logistics Reporting with Automation" {
		t.Errorf("unexpected subject: %s", draft.Subject)
	}

	if draft.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestDraftWithOllama(t *testing.T) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	prospect := Prospect{
		Company:           "Acme Logistics",
		Email:             "contact@acme.example",
		IdentifiedProblem: "Manual logistics reporting",
	}

	draft, err := Draft(context.Background(), client, prospect)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("\nSubject: %s\n\n%s\n", draft.Subject, draft.Body)
}
