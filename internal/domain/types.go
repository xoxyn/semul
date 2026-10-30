package domain

type DeliveryStatus string

const (
    NotSent DeliveryStatus = "not_sent"
    Sending DeliveryStatus = "sending"
    Sent    DeliveryStatus = "sent"
    Failed  DeliveryStatus = "failed"
)

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
