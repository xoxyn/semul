
package batch

import (
	"context"
	"fmt"
	"sync"

	"github.com/xoxyn/semul/internal/domain"
)

type DraftFunc func(
	ctx context.Context,
	sender domain.Sender,
	prospect domain.Prospect,
) (domain.DraftEmail, error)

type result struct {
	index int
	job   domain.EmailJob
	err   error
}

func Process(
	ctx context.Context,
	sender domain.Sender,
	prospects []domain.Prospect,
	draftFn DraftFunc,
	workers int,
) ([]domain.EmailJob, error) {
	if workers < 1 {
		return nil, fmt.Errorf("workers must be at least 1")
	}

	jobs := make([]domain.EmailJob, len(prospects))
	results := make(chan result, len(prospects))

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, prospect := range prospects {
		wg.Add(1)

		go func(i int, prospect domain.Prospect) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if err := ctx.Err(); err != nil {
				results <- result{index: i, err: err}
				return
			}

			draft, err := draftFn(ctx, sender, prospect)
			if err != nil {
				results <- result{
					index: i,
					err:   fmt.Errorf("draft for %s: %w", prospect.Company, err),
				}
				return
			}

			results <- result{
				index: i,
				job: domain.EmailJob{
					Prospect: prospect,
					Draft:    draft,
				},
			}
		}(i, prospect)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error

	for r := range results {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}

		jobs[r.index] = r.job
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return jobs, nil
}
