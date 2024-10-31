package mailer

import (
	"context"
	"crypto/tls"
	epublib "epublib"
	"errors"
	"log"
	netmail "net/mail"
	"os"
	"strconv"
	"time"

	gomail "github.com/go-mail/mail"
)

// MailerService represents a service for managing Mailer.
type MailerService struct {
}

// NewMailerService returns a new instance of MailerService attached to DB.
func NewMailerService() *MailerService {
	return &MailerService{}
}
func (svc *MailerService) SendMail(ctx context.Context, mail epublib.Mail) error {
	port, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("smtp server:", os.Getenv("SMTP_HOST"))
	log.Println("smtp port:", os.Getenv("SMTP_PORT"))
	log.Println("smtp user:", os.Getenv("SMTP_USER"))
	log.Println("smtp pass:", os.Getenv("SMTP_PASS"))
	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), int(port), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	// d.StartTLSPolicy = gomail.MandatoryStartTLS
	d.TLSConfig = &tls.Config{ServerName: os.Getenv("SMTP_HOST")}
	d.Timeout = 5 * time.Second
	d.LocalName = "mailer"

	_, err = netmail.ParseAddress(mail.To)
	if err != nil {
		err := errors.New("invalid-address:" + mail.To)
		log.Println(err)
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", mail.From)
	m.SetHeader("To", mail.To)
	m.SetHeader("Subject", mail.Subject)
	m.SetBody(mail.ContentType, mail.Body)

	log.Println("sending email to ", mail.To)
	err = d.DialAndSend(m)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("sent email to ", mail.To)
	return nil
}
