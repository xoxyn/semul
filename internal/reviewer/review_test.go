package reviewer

import (
    "testing"

    "github.com/xoxyn/semul/internal/domain"
)

func TestEdit(t *testing.T) {
    job := domain.EmailJob{
        Draft: domain.DraftEmail{
            Subject: "Old subject",
            Body:    "Old body",
        },
    }

    Edit(&job, "Updated subject", "Updated body")

    if job.Draft.Subject != "Updated subject" {
        t.Errorf("expected updated subject, got %q", job.Draft.Subject)
    }

    if job.Draft.Body != "Updated body" {
        t.Errorf("expected updated body, got %q", job.Draft.Body)
    }
}

func TestApprove(t *testing.T) {
    job := domain.EmailJob{}

    Approve(&job)

    if !job.Approved {
        t.Error("expected job to be approved")
    }
}
