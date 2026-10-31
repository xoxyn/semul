package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ollama/ollama/api"

	"github.com/xoxyn/semul/internal/batch"
	"github.com/xoxyn/semul/internal/dedupe"
	"github.com/xoxyn/semul/internal/domain"
	"github.com/xoxyn/semul/internal/drafter"
	"github.com/xoxyn/semul/internal/importer"
	"github.com/xoxyn/semul/internal/mailer"
	"github.com/xoxyn/semul/internal/storage"
	"github.com/xoxyn/semul/internal/tui"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("usage: semul prospects.csv")
	}
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	smtp, err := mailer.FromEnv()
	if err != nil {
		return err
	}

	store, err := storage.Open("semul.db")
	if err != nil {
		return err
	}
	defer store.Close()

	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()

	prospects, err := importer.Read(f)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	fresh, skipped, err := dedupe.Filter(ctx, store, prospects)
	if err != nil {
		return err
	}
	for _, p := range skipped {
		fmt.Printf("skip %s (already contacted or duplicate)\n", p.Company)
	}
	if len(fresh) == 0 {
		fmt.Println("Nothing new to contact.")
		return nil
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	sender := domain.Sender{Name: "John", Company: "Semul"}
	draftFn := func(
		ctx context.Context,
		s domain.Sender,
		p domain.Prospect,
	) (domain.DraftEmail, error) {
		return drafter.Draft(ctx, client, s, p)
	}

	fmt.Printf("Drafting %d emails...\n", len(fresh))

	jobs, err := batch.Process(ctx, sender, fresh, draftFn, 2)
	if err != nil {
		return fmt.Errorf("draft emails: %w", err)
	}

	return tui.Run(jobs, dedupe.WrapSend(store, smtp.Send))
}
