package reviewer

import (
	"testing"

	"github.com/xoxyn/semul/internal/domain"
)

func TestEdit(t *testing.T) {
	job := domain.EmailJob{
		Draft: domain.DraftEmail{
			Subject: "old subject",
			Body:    "old body",
		},
	}

	Edit(&job, "updated subject", "updated body")

	if job.Draft.Subject != "updated subject" {
		t.Errorf("expected updated subject, got %q", job.Draft.Subject)
	}

	if job.Draft.Body != "updated body" {
		t.Errorf("expected updated body, got %q", job.Draft.Body)
	}
}

func TestEditResetsApproval(t *testing.T) {
	job := domain.EmailJob{Approved: true}

	Edit(&job, "s", "b")

	if job.Approved {
		t.Error("expected edit to reset approval")
	}
}

func TestApprove(t *testing.T) {
	job := domain.EmailJob{}

	Approve(&job)

	if !job.Approved {
		t.Error("expected job to be approved")
	}
}

func TestUnapprove(t *testing.T) {
	job := domain.EmailJob{Approved: true}

	Unapprove(&job)

	if job.Approved {
		t.Error("expected job to be unapproved")
	}
}

func TestToggle(t *testing.T) {
	job := domain.EmailJob{}

	Toggle(&job)
	if !job.Approved {
		t.Fatal("expected approved after first toggle")
	}

	Toggle(&job)
	if job.Approved {
		t.Fatal("expected unapproved after second toggle")
	}
}
