package batch

import (
	"context"
	"log"

	"github.com/ollama/ollama/api"
	"github.com/xoxyn/semul/internal/domain"
	"github.com/xoxyn/semul/internal/drafter"
)

func Process(p []domain.Prospect,
	ctx context.Context,
	client *api.Client,
	sender domain.Sender,
) []domain.EmailJob {
	var em []domain.EmailJob
	for _, prop := range p {
		draft, err := drafter.Draft(ctx, client, sender, prop)
		if err != nil {
			log.Printf("Couldn't create draft for: %s", prop.Company)
			continue
		}
		emailJob := domain.EmailJob{
			Prospect: prop,
			Draft:    draft,
		}
		em = append(em, emailJob)
	}
	return em
}
