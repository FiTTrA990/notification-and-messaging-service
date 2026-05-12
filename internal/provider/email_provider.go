package provider

import (
	"context"
)

//go:generate mockgen -source=email_provider.go -destination=../../mocks/mock_email_provider.go -package=mocks

// SMTPEmailProvider sends emails via SMTP (stub)
type SMTPEmailProvider struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPEmailProvider creates an SMTP email provider
func NewSMTPEmailProvider(host string, port int, username, password, from string) *SMTPEmailProvider {
	return &SMTPEmailProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

// Send sends an email – STUB
func (p *SMTPEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	// TODO: implement gomail send
	panic("SMTPEmailProvider.Send: not implemented")
}
