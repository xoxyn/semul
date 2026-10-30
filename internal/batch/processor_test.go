package batch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/xoxyn/semul/internal/domain"
	"github.com/xoxyn/semul/internal/drafter"
)

func TestProcess(t *testing.T) {
	sender := domain.Sender{
		Name:    "John",
		Company: "Semul",
	}

	prospects := []domain.Prospect{
		{Company: "Acme", Email: "hello@acme.com", IdentifiedProblem: "Manual reporting"},
		{Company: "Globex", Email: "hello@globex.com", IdentifiedProblem: "Slow workflows"},
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

	jobs, err := Process(context.Background(), sender, prospects, draftFn, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	// Verify original order is preserved.
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

	_, err := Process(context.Background(), domain.Sender{}, prospects, draftFn, 2)

	if err == nil {
		t.Fatal("expected an error")
	}

	if !strings.Contains(err.Error(), "Acme") {
		t.Errorf("expected error to mention Acme, got %v", err)
	}
}

func TestProcessWorkerLimit(t *testing.T) {
	var active atomic.Int32
	var maxActive atomic.Int32

	prospects := make([]domain.Prospect, 6)
	for i := range prospects {
		prospects[i].Company = fmt.Sprintf("Company %d", i)
	}

	draftFn := func(
		ctx context.Context,
		sender domain.Sender,
		prospect domain.Prospect,
	) (domain.DraftEmail, error) {
		current := active.Add(1)
		defer active.Add(-1)

		for {
			max := maxActive.Load()
			if current <= max || maxActive.CompareAndSwap(max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)

		return domain.DraftEmail{
			Subject: "Test",
			Body:    "Test",
		}, nil
	}

	_, err := Process(context.Background(), domain.Sender{}, prospects, draftFn, 2)
	if err != nil {
		t.Fatal(err)
	}

	if maxActive.Load() > 2 {
		t.Errorf("worker limit exceeded: maximum active workers was %d", maxActive.Load())
	}

	if maxActive.Load() < 2 {
		t.Errorf("expected concurrent execution, maximum active workers was %d", maxActive.Load())
	}
}

func TestProcessWithOllama(t *testing.T) {
	if os.Getenv("RUN_OLLAMA_TESTS") != "1" {
		t.Skip("set RUN_OLLAMA_TESTS=1 to run Ollama integration test")
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	sender := domain.Sender{
		Name:    "John",
		Company: "Semul",
	}

	prospects := []domain.Prospect{
		{
			Company:           "Acme Logistics",
			Email:             "contact@acme.example",
			IdentifiedProblem: "Manual logistics reporting",
		},
		{
			Company:           "Globex Solutions",
			Email:             "hello@globex.example",
			IdentifiedProblem: "Repetitive invoice processing",
		},
	}

	draftFn := func(
		ctx context.Context,
		sender domain.Sender,
		prospect domain.Prospect,
	) (domain.DraftEmail, error) {
		return drafter.Draft(ctx, client, sender, prospect)
	}

	jobs, err := Process(ctx, sender, prospects, draftFn, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != len(prospects) {
		t.Fatalf("expected %d jobs, got %d", len(prospects), len(jobs))
	}

	for i, job := range jobs {
		if job.Prospect.Company == "" {
			t.Errorf("job %d has empty prospect company", i)
		}

		if job.Draft.Subject == "" || job.Draft.Body == "" {
			t.Errorf("job %d has an empty email draft", i)
		}

		fmt.Printf(
			"\nCompany: %s\nSubject: %s\n\n%s\n",
			job.Prospect.Company,
			job.Draft.Subject,
			job.Draft.Body,
		)
	}
}
