package drafter

import (
	"context"
	"fmt"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestParseDraft(t *testing.T) {
	input := `Subject: Workflow Automation
Body:
Hi Acme,
We help businesses automate manual workflows.`

	got, err := parseDraft(input)
	if err != nil {
		t.Fatal(err)
	}

	if got.Subject != "Workflow Automation" {
		t.Errorf("got subject %q", got.Subject)
	}

	if got.Body == "" {
		t.Error("expected a non-empty body")
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
