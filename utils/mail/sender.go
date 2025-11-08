package mail

import (
	"fmt"
	"net/smtp"
	"strings"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type Sender interface {
	SendEmail(subject string, body string, to []string, cc []string, bcc []string) error
}

type GmailSender struct {
	name             string
	fromEmailAddress string
	fromEmailPasswd  string
}

func NewGmailSender(name, fromEmailAddress, fromEmailPasswd string) Sender {
	return &GmailSender{
		name:             name,
		fromEmailAddress: fromEmailAddress,
		fromEmailPasswd:  fromEmailPasswd,
	}
}

func (sender *GmailSender) SendEmail(subject string, body string, to []string, cc []string, bcc []string) error {
	//MIME headers
	mime := "MIME-version: 1.0;\n Content-Type: text/html; charset=\"UTF-8\";\n\n"

	//Email headers
	from := fmt.Sprintf("From: %s <%s>\n", sender.name, sender.fromEmailAddress)
	toHeader := fmt.Sprintf("To: %s\r\n", strings.Join(to, ";"))
	ccHeader := fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ";"))
	bccHeader := fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ";"))
	subjectHeader := fmt.Sprintf("Subject: %s\r\n", subject)

	message := []byte(from + toHeader + ccHeader + bccHeader + subjectHeader + mime + body)

	auth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPasswd, "smtp.gmail.com")
	err := smtp.SendMail(smtpServerAddress, auth, sender.fromEmailAddress, to, message)
	if err != nil {
		return fmt.Errorf("smtp.SendMail err: %v", err)
	}
	return nil
}
