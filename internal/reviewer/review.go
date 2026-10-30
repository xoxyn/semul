package reviewer

import "github.com/xoxyn/semul/internal/domain"

// Edit replaces the draft's subject and body. An edited draft is no longer
// what the reviewer approved, so approval is reset.
func Edit(job *domain.EmailJob, subject, body string) {
	job.Draft.Subject = subject
	job.Draft.Body = body
	job.Approved = false
}

func Approve(job *domain.EmailJob) {
	job.Approved = true
}

func Unapprove(job *domain.EmailJob) {
	job.Approved = false
}

// Toggle flips the approval state.
func Toggle(job *domain.EmailJob) {
	job.Approved = !job.Approved
}
