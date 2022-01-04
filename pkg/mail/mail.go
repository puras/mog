package email

import (
	"errors"

	"gopkg.in/gomail.v2"
)

/**
 * @project momo
 * @author <a href="mailto:he@puras.cn">Puras.He</a>
 * @date 2021-08-26 22:49:41
 * @desc 一句话描述功能
 */
type MailServer struct {
	Host     string
	Port     int
	Username string
	Password string
}

type MailInfo struct {
	Subject     string
	FromMail    string
	FromName    string
	ToAddresses []string
	CcMail      string
	CcName      string
	ContentType string
	ContentBody string
	Attachs     []string
}

func SendMail(server MailServer, info MailInfo) error {
	m := gomail.NewMessage()
	if info.FromMail == "" {
		return errors.New("From mail required")
	}
	if len(info.ToAddresses) == 0 && info.CcMail == "" {
		return errors.New("To mail or Cc mail required")
	}
	m.SetHeader("From", info.FromMail, info.FromName)
	if len(info.ToAddresses) > 0 {
		m.SetHeader("To", info.ToAddresses...)
	}
	if info.CcMail != "" {
		m.SetAddressHeader("Cc", info.CcMail, info.CcName)
	}
	if info.Subject != "" {
		m.SetHeader("Subject", info.Subject)
	}
	if info.ContentBody != "" {
		if info.ContentType == "" {
			info.ContentType = "text/html"
		}
		m.SetBody(info.ContentType, info.ContentBody)
	}
	if len(info.Attachs) > 0 {
		for _, attach := range info.Attachs {
			m.Attach(attach)
		}
	}

	d := gomail.NewDialer(server.Host, server.Port, server.Username, server.Password)
	return d.DialAndSend(m)
}
