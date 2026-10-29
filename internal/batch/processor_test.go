package batch

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xoxyn/semul/internal/domain"
)

func TestProcess(t *testing.T) {
	sender := domain.Sender{
		Name:    "John",
		Company: "Semul",
	}

	prospects := []domain.Prospect{
		{
			Company:           "Acme",
			Email:             "hello@acme.com",
			IdentifiedProblem: "Manual reporting",
		},
		{
			Company:           "Globex",
			Email:             "hello@globex.com",
			IdentifiedProblem: "Slow workflows",
		},
	}

	draftFn := func(
		ctx context.Context,
		sender domain.Sender,
		prospect domain.Prospect,
	) (domain.DraftEmail, error) {
		return domain.DraftEmail{
			Subject: "Automation for " + prospect.Company,
			Body:    "Hello " + prospect.Company,
		}, nil
	}

	jobs, err := Process(context.Background(), sender, prospects, draftFn)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	if jobs[0].Prospect.Company != "Acme" {
		t.Errorf("expected Acme, got %s", jobs[0].Prospect.Company)
	}

	if jobs[0].Draft.Subject != "Automation for Acme" {
		t.Errorf("unexpected subject: %s", jobs[0].Draft.Subject)
	}

	if jobs[1].Prospect.Company != "Globex" {
		t.Errorf("expected Globex, got %s", jobs[1].Prospect.Company)
	}
}

func TestProcessDraftError(t *testing.T) {
	prospects := []domain.Prospect{
		{Company: "Acme"},
	}

	draftFn := func(
		ctx context.Context,
		sender domain.Sender,
		prospect domain.Prospect,
	) (domain.DraftEmail, error) {
		return domain.DraftEmail{}, errors.New("generation failed")
	}

	_, err := Process(context.Background(), domain.Sender{}, prospects, draftFn)

	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "Acme") {
		t.Errorf("expected error to mention Acme, got %v", err)
	}
}
