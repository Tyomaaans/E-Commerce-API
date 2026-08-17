package users

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"

	"E-COMMERCE-API/internal/config"
)

// ==== Data Payload Template ====

// EmailData menampung data opsional/dinamis yang dikirim ke template email.
type EmailData struct {
	Name    string
	Link    string
	Amount  string
	ItemName string
	// Tambahkan field umum lainnya sesuai kebutuhan (invoice, promo, dll)
}

// ==== Service Interface ====

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, link string) error
	SendRegisterSuccessEmail(ctx context.Context, to, name string) error
	SendResetPasswordEmail(ctx context.Context, to, link string) error
	SendUpdatePasswordSuccessEmail(ctx context.Context, to, name string) error
	SendInvoiceEmail(ctx context.Context, to string, data EmailData) error
}

// ==== Implementation ====

type smtpEmailService struct {
	host     string
	port     string
	from     string
	username string
	password string
}

func NewEmailService(cfg config.AppConfig) EmailService {
	return &smtpEmailService{
		host:     cfg.SMTPhost,
		port:     cfg.SMTPport,
		from:     cfg.SMTPfrom,
		username: cfg.SMTPusername,
		password: cfg.SMTPpassword,
	}
}

// ==== High-Level Methods ====

func (s *smtpEmailService) SendVerificationEmail(ctx context.Context, to, link string) error {
	return s.sendTemplatedEmail(to, "Verify Your Email", verifyEmailTpl, EmailData{Link: link})
}

func (s *smtpEmailService) SendRegisterSuccessEmail(ctx context.Context, to, name string) error {
	return s.sendTemplatedEmail(to, "Welcome to Our Platform!", registerSuccessTpl, EmailData{Name: name})
}

func (s *smtpEmailService) SendResetPasswordEmail(ctx context.Context, to, link string) error {
	return s.sendTemplatedEmail(to, "Reset Your Password", resetPasswordTpl, EmailData{Link: link})
}

func (s *smtpEmailService) SendUpdatePasswordSuccessEmail(ctx context.Context, to, name string) error {
	return s.sendTemplatedEmail(to, "Your Password Has Been Updated", updatePasswordSuccessTpl, EmailData{Name: name})
}

func (s *smtpEmailService) SendInvoiceEmail(ctx context.Context, to string, data EmailData) error {
	return s.sendTemplatedEmail(to, "Your Payment Receipt", invoiceTpl, data)
}

// ==== Generic Reusable Helper ====

// sendTemplatedEmail merender template HTML dan mengirimkannya via SMTP.
func (s *smtpEmailService) sendTemplatedEmail(to, subject, tplSource string, data EmailData) error {
	body, err := renderTemplate(tplSource, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return s.send(to, subject, body)
}

func (s *smtpEmailService) send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.from, to, subject,
	)
	msg := []byte(headers + htmlBody)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}

func renderTemplate(tplSource string, data any) (string, error) {
	t, err := template.New("email").Parse(tplSource)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ==== Email HTML Templates ====

const verifyEmailTpl = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 20px;">
  <div style="max-width: 520px; margin: auto; background: #fff; padding: 32px; border-radius: 8px;">
    <h2 style="color: #333;">Verify Your Email</h2>
    <p style="color: #555;">
      Thank you for registering! Click the button below to verify your email address.
      This link will expire in <strong>24 hours</strong>.
    </p>
    <a href="{{.Link}}" style="display: inline-block; padding: 12px 24px; background: #4F46E5; color: #fff; text-decoration: none; border-radius: 6px; margin-top: 16px;">Verify Email</a>
  </div>
</body>
</html>`

const registerSuccessTpl = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 20px;">
  <div style="max-width: 520px; margin: auto; background: #fff; padding: 32px; border-radius: 8px;">
    <h2 style="color: #333;">Welcome, {{.Name}}! 🎉</h2>
    <p style="color: #555;">
      Your email has been verified successfully. Your account is now active and ready to use!
    </p>
  </div>
</body>
</html>`

const resetPasswordTpl = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 20px;">
  <div style="max-width: 520px; margin: auto; background: #fff; padding: 32px; border-radius: 8px;">
    <h2 style="color: #333;">Reset Your Password</h2>
    <p style="color: #555;">
      We received a request to reset your password. Click the button below.
      This link will expire in <strong>15 minutes</strong>.
    </p>
    <a href="{{.Link}}" style="display: inline-block; padding: 12px 24px; background: #DC2626; color: #fff; text-decoration: none; border-radius: 6px; margin-top: 16px;">Reset Password</a>
  </div>
</body>
</html>`

const updatePasswordSuccessTpl = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 20px;">
  <div style="max-width: 520px; margin: auto; background: #fff; padding: 32px; border-radius: 8px;">
    <h2 style="color: #333;">Password Updated</h2>
    <p style="color: #555;">
      Hello {{.Name}}, your password was successfully changed. If you did not initiate this change, please contact our support immediately.
    </p>
  </div>
</body>
</html>`

const invoiceTpl = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 20px;">
  <div style="max-width: 520px; margin: auto; background: #fff; padding: 32px; border-radius: 8px;">
    <h2 style="color: #333;">Payment Confirmation</h2>
    <p style="color: #555;">Hi {{.Name}}, thanks for your purchase!</p>
    <div style="background: #f9fafb; padding: 16px; border-radius: 6px; margin: 16px 0;">
      <p style="margin: 4px 0;"><strong>Item:</strong> {{.ItemName}}</p>
      <p style="margin: 4px 0;"><strong>Total Paid:</strong> {{.Amount}}</p>
    </div>
  </div>
</body>
</html>`