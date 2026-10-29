package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wneessen/go-mail"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	message := mail.NewMsg()

	if err := message.From("numterminal@gmail.com"); err != nil {
		log.Fatalf("failed to set From: %s", err)
	}

	if err := message.To("numterminal2@gmail.com"); err != nil {
		log.Fatalf("failed to set To: %s", err)
	}

	message.Subject("My first email with Go")
	message.SetBodyString(
		mail.TypeTextPlain,
		"Hello! This email was sent using Go and Gmail SMTP from semul.",
	)

	client, err := mail.NewClient(
		"smtp.gmail.com",
		mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(os.Getenv("SMTP_USERNAME")),
		mail.WithPassword(os.Getenv("SMTP_PASSWORD")),
	)
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	if err := client.DialAndSend(message); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}

	log.Println("Email sent successfully!")
}
