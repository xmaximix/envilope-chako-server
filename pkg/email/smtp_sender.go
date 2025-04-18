package email

import (
	"context"
	"fmt"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"net/smtp"
	"time"
)

type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type SMTPSender struct {
	addr string
	from string
	auth smtp.Auth
}

func NewSMTPSender(host string, port int, user, pass, from string) *SMTPSender {
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", user, pass, host)
	return &SMTPSender{addr: addr, from: from, auth: auth}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s",
		s.from, to, subject, body,
	)

	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(s.addr, s.auth, s.from, []string{to}, []byte(msg))
	}()
	select {
	case err := <-done:
		return errs.Wrap("smtp send", err)
	case <-ctx.Done():
		return errs.Wrap("email send canceled", ctx.Err())
	case <-time.After(5 * time.Second):
		return errs.Wrap("email send timeout", ctx.Err())
	}
}
