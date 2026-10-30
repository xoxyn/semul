package domain

type Sender struct {
	Name    string
	Company string
}

type Prospect struct {
	Company           string
	Email             string
	IdentifiedProblem string
}

type DraftEmail struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailJob struct {
	Prospect
	Draft DraftEmail

	Approved bool
	Status   DeliveryStatus
}
