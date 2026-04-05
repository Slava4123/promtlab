package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"
	"time"

	"promptvault/internal/infrastructure/config"
)

type Service struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewService(cfg *config.SMTPConfig) *Service {
	return &Service{
		host: cfg.Host,
		port: cfg.Port,
		user: cfg.User,
		pass: cfg.Password,
		from: cfg.From,
	}
}

func (s *Service) Configured() bool {
	return s.host != "" && s.user != "" && s.pass != ""
}

// --- Public API ---

func (s *Service) SendVerificationCode(to, code string) error {
	return s.send(to,
		"Подтверждение email — ПромтЛаб",
		fmt.Sprintf("Ваш код подтверждения: %s\r\n\r\nКод действителен 15 минут.\r\n\r\nЕсли вы не регистрировались в ПромтЛаб, проигнорируйте это письмо.", code),
	)
}

func (s *Service) SendPasswordResetCode(to, code string) error {
	return s.send(to,
		"Сброс пароля — ПромтЛаб",
		fmt.Sprintf("Код для сброса пароля: %s\r\n\r\nКод действителен 15 минут.\r\n\r\nЕсли вы не запрашивали сброс пароля, проигнорируйте это письмо.", code),
	)
}

func (s *Service) SendSetPasswordCode(to, code string) error {
	return s.send(to,
		"Установка пароля — ПромтЛаб",
		fmt.Sprintf("Код для установки пароля: %s\r\n\r\nКод действителен 15 минут.\r\n\r\nЕсли вы не запрашивали установку пароля, проигнорируйте это письмо.", code),
	)
}

func (s *Service) SendPasswordChangedNotification(to string) error {
	return s.send(to,
		"Пароль изменён — ПромтЛаб",
		"Ваш пароль в ПромтЛаб был изменён.\r\n\r\nЕсли это были не вы, немедленно войдите в аккаунт и смените пароль или свяжитесь с поддержкой.",
	)
}

// --- Internal ---

func (s *Service) send(to, subject, body string) error {
	msg := s.buildMessage(to, subject, body)

	var lastErr error
	for attempt := range 3 {
		if s.port == 465 {
			lastErr = s.sendSMTPS(to, msg)
		} else {
			lastErr = s.sendSTARTTLS(to, msg)
		}
		if lastErr == nil {
			return nil
		}
		slog.Warn("email send failed, retrying", "attempt", attempt+1, "to", to, "error", lastErr)
		time.Sleep(time.Duration(1<<attempt) * time.Second)
	}
	return fmt.Errorf("email send failed after 3 attempts: %w", lastErr)
}

func (s *Service) buildMessage(to, subject, body string) []byte {
	fromEncoded := fmt.Sprintf("=?utf-8?B?%s?= <%s>", base64.StdEncoding.EncodeToString([]byte("ПромтЛаб")), s.from)
	subjectEncoded := fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		fromEncoded, to, subjectEncoded, body)
	return []byte(msg)
}

// sendSMTPS — порт 465, TLS с самого начала (для Docker Desktop на Windows)
func (s *Service) sendSMTPS(to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Auth(smtp.PlainAuth("", s.user, s.pass, s.host)); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return client.Quit()
}

// sendSTARTTLS — порт 587, стандартный smtp.SendMail (для production/VPS)
func (s *Service) sendSTARTTLS(to string, msg []byte) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
