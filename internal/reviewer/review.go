package reviewer

import "github.com/xoxyn/semul/internal/domain"

func Edit(job *domain.EmailJob, subject, body string) {
    job.Draft.Subject = subject
    job.Draft.Body = body
}

func Approve(job *domain.EmailJob) {
    job.Approved = true
}
