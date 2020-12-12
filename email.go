package notifier

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/jordan-wright/email"
)

type EmailNotifier struct {
	sender    string
	tlsConfig *tls.Config
	smtpAuth  smtp.Auth
}

func NewEmailNotifier(server, username, password string) (Notifier, error) {
	if server == "" {
		return nil, errors.New("server is required")
	}

	if username == "" {
		return nil, errors.New("username is required")
	}

	if password == "" {
		return nil, errors.New("password is required")
	}

	emailNotifier := EmailNotifier{
		sender: username,
		tlsConfig: &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         server,
		},
	}

	emailNotifier.smtpAuth = smtp.PlainAuth("", emailNotifier.sender, password, emailNotifier.tlsConfig.ServerName)

	return emailNotifier, nil

}

func (emn EmailNotifier) Notify(receivers []string, subject, content string, attachments ...Attachment) error {
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

	return em.SendWithTLS(emn.tlsConfig.ServerName, emn.smtpAuth, emn.tlsConfig)
}
