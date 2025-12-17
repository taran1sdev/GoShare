package test

import (
	"fmt"
	"os"
	"testing"

	"taran1s.share/models"
)

func TestEmail(t *testing.T) {
	email := models.Email{
		To:        "thomasaird0@gmail.com",
		From:      "support@taran1s.me",
		Subject:   "Test email",
		Plaintext: "This is the body",
		HTML:      `<h1>This is a test!</h1><p>this is the body</p><p>enjoy</p>`,
	}

	es := models.NewEmailService(models.SMTPConfig{
		Host:     os.Getenv("EMAILHOST"),
		Port:     25,
		Username: os.Getenv("EMAILUSERNAME"),
		Password: os.Getenv("EMAILPASSWORD"),
	})

	err := es.Send(email)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Email sent! Check your inbox...")
}
