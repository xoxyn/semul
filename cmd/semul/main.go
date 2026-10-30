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
	"github.com/xoxyn/semul/internal/domain"
	"github.com/xoxyn/semul/internal/drafter"
	"github.com/xoxyn/semul/internal/importer"
	"github.com/xoxyn/semul/internal/mailer"
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

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	smtp, err := mailer.FromEnv()
	if err != nil {
		return err
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()

	prospects, err := importer.Read(f)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
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

	fmt.Printf("Drafting %d emails...\n", len(prospects))

	jobs, err := batch.Process(context.Background(), sender, prospects, draftFn, 2)
	if err != nil {
		return fmt.Errorf("draft emails: %w", err)
	}

	return tui.Run(jobs, smtp.Send)
}
