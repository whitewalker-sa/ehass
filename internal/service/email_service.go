package service

import (
	"context"
	"fmt"
	"net/smtp"
)

// emailService implements EmailService interface
type emailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	appBaseURL   string
}

// NewEmailService creates a new email service
func NewEmailService(
	smtpHost string,
	smtpPort int,
	smtpUsername string,
	smtpPassword string,
	fromEmail string,
	appBaseURL string,
) EmailService {
	return &emailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		appBaseURL:   appBaseURL,
	}
}

// SendVerificationEmail sends an email with a verification link
func (s *emailService) SendVerificationEmail(ctx context.Context, email, name, token string) error {
	subject := "Verify Your Email Address"
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.appBaseURL, token)

	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Email Verification</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; 
				text-decoration: none; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h2>Welcome to EHASS, %s!</h2>
			<p>Thank you for registering with us. Please verify your email address by clicking the button below:</p>
			<p><a href="%s" class="button">Verify Email</a></p>
			<p>Or copy and paste this link in your browser:</p>
			<p>%s</p>
			<p>If you didn't register for an account, you can safely ignore this email.</p>
			<p>Best regards,<br>The EHASS Team</p>
		</div>
	</body>
	</html>
	`, name, verificationLink, verificationLink)

	return s.sendEmail(email, subject, body)
}

// SendPasswordResetEmail sends an email with password reset link
func (s *emailService) SendPasswordResetEmail(ctx context.Context, email, name, token string) error {
	subject := "Reset Your Password"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.appBaseURL, token)

	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Password Reset</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; 
				text-decoration: none; border-radius: 5px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h2>Hello, %s!</h2>
			<p>We received a request to reset your password. If you didn't make this request, you can safely ignore this email.</p>
			<p>To reset your password, click the button below:</p>
			<p><a href="%s" class="button">Reset Password</a></p>
			<p>Or copy and paste this link in your browser:</p>
			<p>%s</p>
			<p>This link will expire in 1 hour for security reasons.</p>
			<p>Best regards,<br>The EHASS Team</p>
		</div>
	</body>
	</html>
	`, name, resetLink, resetLink)

	return s.sendEmail(email, subject, body)
}

// sendEmail sends an email using SMTP
func (s *emailService) sendEmail(to, subject, body string) error {
	// Set up authentication information
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Construct email headers and body
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("Subject: " + subject + "\r\n" +
		"From: " + s.fromEmail + "\r\n" +
		"To: " + to + "\r\n" +
		mime + "\r\n" +
		body)

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
}
