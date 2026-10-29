package batch

import (
	"context"
	"fmt"

	"github.com/xoxyn/semul/internal/domain"
)

type DraftFunc func(
	ctx context.Context,
	sender domain.Sender,
	prospect domain.Prospect,
) (domain.DraftEmail, error)

func Process(
	ctx context.Context,
	sender domain.Sender,
	prospects []domain.Prospect,
	draftFn DraftFunc,
) ([]domain.EmailJob, error) {
	var jobs []domain.EmailJob

	for _, prospect := range prospects {
		draft, err := draftFn(ctx, sender, prospect)
		if err != nil {
			return nil, fmt.Errorf("draft for %s: %w", prospect.Company, err)
		}

		jobs = append(jobs, domain.EmailJob{
			Prospect: prospect,
			Draft:    draft,
		})
	}

	return jobs, nil
}
