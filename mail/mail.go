package mail

import (
	"context"
	"gopkg.in/gomail.v2"
	"sync"
	"time"
)

var (
	globalSender *SmtpSender
	once         sync.Once
)

func SetSender(sender *SmtpSender) {
	once.Do(func() {
		globalSender = sender
	})
}

func Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	return globalSender.Send(ctx, to, cc, bcc, subject, body, file...)
}

func SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	return globalSender.SendTo(ctx, to, subject, body, file...)
}

type SmtpSender struct {
	SmtpHost string
	Port     int
	FromName string
	FromMail string
	UserName string
	AuthCode string
}

func (o *SmtpSender) Send(ctx context.Context, to []string, cc []string, bcc []string, subject string, body string, file ...string) error {
	msg := gomail.NewMessage(gomail.SetEncoding(gomail.Base64))
	msg.SetHeader("From", msg.FormatAddress(o.FromName, o.FromName))
	msg.SetHeader("To", to...)
	msg.SetHeader("Cc", cc...)
	msg.SetHeader("Bcc", bcc...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html;charset=utf-8", body)

	for _, v := range file {
		msg.Attach(v)
	}
	d := gomail.NewDialer(o.SmtpHost, o.Port, o.UserName, o.AuthCode)
	return d.DialAndSend(msg)
}

func (o *SmtpSender) SendTo(ctx context.Context, to []string, subject string, body string, file ...string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = o.Send(ctx, to, nil, nil, subject, body, file...)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		err = nil
		break
	}
	return err
}
