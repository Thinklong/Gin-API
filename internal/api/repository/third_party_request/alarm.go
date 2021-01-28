package third_party_request

import (
	"gin-api/pkg/httpclient"
	"gin-api/pkg/mail"
)

// 实现 AlarmObject 告警
var _ httpclient.AlarmObject = (*AlarmEmail)(nil)

type AlarmEmail struct{}

func (a *AlarmEmail) Send(subject, body string) error {
	options := &mail.Options{
		MailHost: "smtp.163.com",
		MailPort: 465,
		MailUser: "xx@163.com",
		MailPass: "",
		MailTo:   "",
		Subject:  subject,
		Body:     body,
	}
	return mail.Send(options)
}
