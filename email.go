package notifier

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/jordan-wright/email"
)

type EMailNotifier struct {
	sender     string
	tlsAddress string
	tlsConfig  *tls.Config
	smtpAuth   smtp.Auth
}

func CreateEMailNotifier() (Notifier, error) {
	if err := checkEnv(
		"NOTIFIER_EMAIL_TLS_ADDRESS",
		"NOTIFIER_EMAIL_TLS_PORT",
		"NOTIFIER_EMAIL_SENDER_USERNAME",
		"NOTIFIER_EMAIL_SENDER_PASSWORD"); err != nil {
		return nil, err
	}

	tlsAddress := os.Getenv("NOTIFIER_EMAIL_TLS_ADDRESS")
	if port := os.Getenv("NOTIFIER_EMAIL_TLS_PORT"); port != "" {
		tlsAddress += ":" + port
	}

	emailNotifier := EMailNotifier{
		sender:     os.Getenv("NOTIFIER_EMAIL_SENDER_USERNAME"),
		tlsAddress: tlsAddress,
		tlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         tlsAddress,
		},
	}

	emailNotifier.smtpAuth = smtp.PlainAuth("", emailNotifier.sender, os.Getenv("NOTIFIER_EMAIL_SENDER_PASSWORD"), tlsAddress)

	return emailNotifier, nil

}

func (emn EMailNotifier) Notify(receivers []string, subject, content string, attachments ...Attachment) error {
	em := email.NewEmail()
	em.From = emn.sender
	em.Subject = subject
	em.HTML = []byte(content)
	to, cc := []string{}, []string{}

	for _, rcvr := range receivers {
		arr := strings.Split(rcvr, ":")
		if len(arr) < 2 {
			continue
		}

		prefix, receiver := strings.ToLower(arr[0]), arr[1]
		if prefix == "to" {
			to = append(to, receiver)
		} else if prefix == "cc" {
			cc = append(cc, receiver)
		}
	}

	em.To = to
	em.Cc = cc

	if len(attachments) > 0 {
		for _, att := range attachments {
			r := bytes.NewReader(att.Content)
			_, err := em.Attach(r, att.FileName, att.ContentType)
			if err != nil {
				return fmt.Errorf("attach file(%s) failed, error: %s", att.FileName, err.Error())
			}
		}
	}

	return em.SendWithTLS(emn.tlsAddress, emn.smtpAuth, emn.tlsConfig)
}

func checkEnv(keys ...string) error {
	for _, key := range keys {
		if os.Getenv(key) == "" {
			return fmt.Errorf("env %s is required", key)
		}
	}

	return nil
}
