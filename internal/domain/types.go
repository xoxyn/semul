// Package Domain Contains global types
package domain

import (
	"errors"
	"strings"
)

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


// CompanyKey identifies a company regardless of case and spacing, so
// "Acme  Logistics" and "acme logistics" count as the same company.
func (p Prospect) CompanyKey() string {
	return strings.Join(strings.Fields(strings.ToLower(p.Company)), " ")
}

// EmailKey identifies an address regardless of case and surrounding spaces.
func (p Prospect) EmailKey() string {
	return strings.ToLower(strings.TrimSpace(p.Email))
}

// ErrAlreadyContacted is returned by a sender that refuses to email a
// prospect we have already emailed.
var ErrAlreadyContacted = errors.New("already contacted")
