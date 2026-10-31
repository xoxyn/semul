// Package dedupe keeps us from emailing the same prospect twice.
//
// It works in two places:
//
//  1. Filter runs right after the CSV is imported, so already-contacted
//     prospects are never drafted (saving LLM time) or shown for review.
//  2. WrapSend wraps the real sender so that, at the moment of sending, we
//     re-check history and record every successful send.
package dedupe

import (
	"context"
	"fmt"

	"github.com/xoxyn/semul/internal/domain"
)

// Store is the persistence this package needs. internal/storage implements it.
type Store interface {
	// Contacted reports whether this prospect's company or email has
	// already been emailed.
	Contacted(ctx context.Context, p domain.Prospect) (bool, error)
	// MarkSent records a successful send.
	MarkSent(ctx context.Context, job domain.EmailJob) error
}

// Filter splits prospects into new ones and ones to skip. A prospect is
// skipped if it was contacted in a previous run, or if the same company or
// address appears earlier in this same list.
func Filter(
	ctx context.Context,
	store Store,
	prospects []domain.Prospect,
) (fresh, skipped []domain.Prospect, err error) {
	seenCompany := map[string]bool{}
	seenEmail := map[string]bool{}

	for _, p := range prospects {
		ck, ek := p.CompanyKey(), p.EmailKey()

		if seenCompany[ck] || seenEmail[ek] {
			skipped = append(skipped, p)
			continue
		}
		seenCompany[ck], seenEmail[ek] = true, true

		contacted, err := store.Contacted(ctx, p)
		if err != nil {
			return nil, nil, fmt.Errorf("check history for %s: %w", p.Company, err)
		}
		if contacted {
			skipped = append(skipped, p)
			continue
		}
		fresh = append(fresh, p)
	}
	return fresh, skipped, nil
}

// WrapSend returns a sender that refuses duplicates and records successes.
// It has the same signature as tui.SendFunc.
func WrapSend(
	store Store,
	send func(ctx context.Context, job domain.EmailJob) error,
) func(ctx context.Context, job domain.EmailJob) error {
	return func(ctx context.Context, job domain.EmailJob) error {
		contacted, err := store.Contacted(ctx, job.Prospect)
		if err != nil {
			// Fail closed: if we can't check history, don't send.
			return fmt.Errorf("check history: %w", err)
		}
		if contacted {
			return domain.ErrAlreadyContacted
		}

		if err := send(ctx, job); err != nil {
			return err // not recorded, so it can be retried later
		}

		if err := store.MarkSent(ctx, job); err != nil {
			return fmt.Errorf("email WAS sent but could not be recorded: %w", err)
		}
		return nil
	}
}
