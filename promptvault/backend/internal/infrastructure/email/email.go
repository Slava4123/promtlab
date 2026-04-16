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

// SendWelcome — приветственное письмо после подтверждения email.
// Дружеский тон на "ты". Знакомит с 3 ключевыми фичами чтобы юзер сразу увидел
// ценность и не забросил аккаунт после регистрации (M-5: D1-retention lift).
// frontendURL — корень приложения для CTA-ссылок.
func (s *Service) SendWelcome(to, name, frontendURL string) error {
	greeting := "Привет"
	if name != "" {
		greeting = fmt.Sprintf("Привет, %s", name)
	}
	body := fmt.Sprintf(
		"%s!\r\n\r\n"+
			"Спасибо, что присоединился к ПромтЛаб — месту, где твои AI-промпты перестают теряться в заметках.\r\n\r\n"+
			"Вот с чего стоит начать:\r\n\r\n"+
			"• Создай первый промпт — %s/prompts/new\r\n"+
			"• Попробуй AI-улучшение — внутри редактора есть кнопка «Улучшить через AI» — Claude сделает твой промпт конкретнее\r\n"+
			"• Подключи Chrome-расширение — вставляй промпты в ChatGPT/Claude/Gemini в два клика\r\n\r\n"+
			"Если что-то непонятно — пиши на %s, отвечу лично.\r\n\r\n"+
			"Слава, создатель ПромтЛаб",
		greeting, frontendURL, supportEmail,
	)
	return s.send(to, "Добро пожаловать в ПромтЛаб 👋", body)
}

// supportEmail — адрес поддержки для welcome/transactional писем. Pinning на константу
// чтобы не требовать новую конфигурацию и не раскидывать по коду.
const supportEmail = "slava0gpt@gmail.com"

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

func (s *Service) SendTeamInvitation(to, teamName, inviterName string) error {
	return s.send(to,
		fmt.Sprintf("Приглашение в команду «%s» — ПромтЛаб", teamName),
		fmt.Sprintf("%s приглашает вас в команду «%s» на ПромтЛаб.\r\n\r\nВойдите в приложение, чтобы принять или отклонить приглашение.", inviterName, teamName),
	)
}

// SendRenewalFailed уведомляет юзера о неудачной попытке автопродления.
// Причины обычно: недостаточно средств, карта истекла, банк-эмитент отклонил.
// attempt/maxAttempts — номер попытки для информативного текста
// («попытка 1 из 3» снижает тревожность, «последняя попытка» призывает к действию).
// graceUntil — если задано, после исчерпания retry доступ сохраняется до этой даты (M-9).
func (s *Service) SendRenewalFailed(to, planName string, attempt, maxAttempts int, endsAt time.Time, graceUntil *time.Time, frontendURL string) error {
	var body string
	subject := fmt.Sprintf("Не удалось продлить подписку %s — ПромтЛаб", planName)

	switch {
	case graceUntil != nil:
		// Последняя попытка провалилась — grace period до downgrade.
		subject = fmt.Sprintf("Подписка %s: требуется действие — ПромтЛаб", planName)
		body = fmt.Sprintf(
			"Мы сделали %d попытки списания за подписку ПромтЛаб %s — все неуспешные.\r\n\r\n"+
				"Ваш доступ сохраняется до %s. После этой даты аккаунт перейдёт на Free план, если карта не будет обновлена.\r\n\r\n"+
				"Обновить способ оплаты: %s/settings\r\n\r\n"+
				"Созданные промпты и коллекции при переходе на Free сохранятся — но часть возможностей ограничится.",
			maxAttempts, planName, graceUntil.Format("02.01.2006"), frontendURL,
		)
	case attempt >= maxAttempts:
		body = fmt.Sprintf(
			"Последняя попытка списания за подписку ПромтЛаб %s не удалась (%d из %d).\r\n\r\n"+
				"Возможные причины: недостаточно средств, карта истекла или банк отклонил списание.\r\n\r\n"+
				"Подписка действует до %s. Обновите способ оплаты, чтобы не потерять доступ: %s/settings",
			planName, attempt, maxAttempts, endsAt.Format("02.01.2006"), frontendURL,
		)
	default:
		body = fmt.Sprintf(
			"Не удалось продлить подписку ПромтЛаб %s (попытка %d из %d).\r\n\r\n"+
				"Возможные причины: недостаточно средств, карта истекла или банк отклонил списание.\r\n\r\n"+
				"Подписка остаётся активной до %s. Мы автоматически попробуем списать ещё раз через 24 часа.\r\n\r\n"+
				"Обновить способ оплаты можно в настройках: %s/settings",
			planName, attempt, maxAttempts, endsAt.Format("02.01.2006"), frontendURL,
		)
	}
	return s.send(to, subject, body)
}

// SendReengagement — письмо для юзеров, не заходивших 14+ дней (M-5d).
// Напоминает о ценности, не давит на продажу. Один раз в 30 дней на юзера.
func (s *Service) SendReengagement(to, name, frontendURL string) error {
	greeting := "Привет"
	if name != "" {
		greeting = fmt.Sprintf("Привет, %s", name)
	}
	body := fmt.Sprintf(
		"%s!\r\n\r\n"+
			"Давно не видели тебя в ПромтЛаб. Всё в порядке?\r\n\r\n"+
			"Если есть хотя бы один полезный промпт, который ты часто используешь — держи ссылку на библиотеку:\r\n%s/dashboard\r\n\r\n"+
			"А если помнишь — у нас есть:\r\n"+
			"• AI-улучшение промптов одним кликом\r\n"+
			"• Chrome-расширение для вставки в ChatGPT\r\n"+
			"• MCP-сервер — работа с промптами прямо из Claude Desktop\r\n\r\n"+
			"Если что-то не устраивает — напиши на %s, разберусь лично. Если отписка — ответь этим письмом с «unsubscribe».",
		greeting, frontendURL, supportEmail,
	)
	return s.send(to, "Давно не видели тебя — ПромтЛаб", body)
}

// SendQuotaWarning — предупреждение когда юзер достиг 80% квоты (M-5c).
// quotaType: "ai_total" (Free — одноразово) / "ai_daily" (Pro/Max — сегодня).
// used/limit — текущий счётчик, frontendURL — для CTA на /pricing.
func (s *Service) SendQuotaWarning(to, name, quotaType string, used, limit int, frontendURL string) error {
	greeting := "Привет"
	if name != "" {
		greeting = fmt.Sprintf("Привет, %s", name)
	}
	var subject, body string
	switch quotaType {
	case "ai_total":
		// Free: 4 из 5 запросов за всю жизнь аккаунта — жёсткий апсейл-момент.
		subject = "Остался 1 AI-запрос — ПромтЛаб"
		body = fmt.Sprintf(
			"%s!\r\n\r\n"+
				"Ты использовал %d из %d AI-запросов на Free. Остался всего %d — потом AI-улучшение промптов станет недоступно.\r\n\r\n"+
				"На Pro — 10 запросов каждый день (300 в месяц), за 599₽ в месяц — меньше 20₽ в день.\r\n\r\n"+
				"Оформить: %s/pricing",
			greeting, used, limit, limit-used, frontendURL,
		)
	default: // ai_daily
		subject = "80% дневной квоты AI исчерпано — ПромтЛаб"
		body = fmt.Sprintf(
			"%s!\r\n\r\n"+
				"Ты уже использовал %d из %d AI-запросов сегодня — осталось %d.\r\n\r\n"+
				"Если лимит заканчивается слишком быстро — посмотри Max: 15 запросов в день, 13990₽ в год.\r\n\r\n"+
				"Посмотреть: %s/pricing",
			greeting, used, limit, limit-used, frontendURL,
		)
	}
	return s.send(to, subject, body)
}

// SendPreExpireReminder — напоминание о скором окончании подписки (M-5b).
// Отправляется из ReminderLoop за 3 и 1 день до period_end для юзеров с
// auto_renew=false (ручное продление). daysLeft — 3 или 1, от этого зависит
// тон письма ("продли заранее" vs "последний шанс").
func (s *Service) SendPreExpireReminder(to, planName string, daysLeft int, endsAt time.Time, frontendURL string) error {
	var subject, body string
	switch {
	case daysLeft <= 1:
		subject = fmt.Sprintf("Подписка %s истекает завтра — ПромтЛаб", planName)
		body = fmt.Sprintf(
			"Твоя подписка ПромтЛаб %s истекает уже завтра (%s).\r\n\r\n"+
				"После этой даты аккаунт перейдёт на Free — 50 промптов, 3 коллекции, 5 AI-запросов всего.\r\n\r\n"+
				"Чтобы сохранить доступ — продли подписку: %s/pricing\r\n\r\n"+
				"Все промпты останутся на месте, просто часть возможностей станет ограничена.",
			planName, endsAt.Format("02.01.2006"), frontendURL,
		)
	default:
		subject = fmt.Sprintf("Подписка %s истекает через %d дня — ПромтЛаб", planName, daysLeft)
		body = fmt.Sprintf(
			"Привет!\r\n\r\n"+
				"Твоя подписка ПромтЛаб %s истекает через %d дня (%s).\r\n\r\n"+
				"Если хочешь сохранить доступ — продли заранее: %s/pricing\r\n\r\n"+
				"Если нет — не переживай, промпты не удалятся, просто переключимся на Free лимиты.",
			planName, daysLeft, endsAt.Format("02.01.2006"), frontendURL,
		)
	}
	return s.send(to, subject, body)
}

// SendSubscriptionExpired уведомляет о переводе на Free после исчерпания
// retry-попыток. Отправляется из ExpirationLoop когда подписка переходит в expired.
func (s *Service) SendSubscriptionExpired(to, planName, frontendURL string) error {
	body := fmt.Sprintf(
		"Подписка ПромтЛаб %s истекла после нескольких неудачных попыток автопродления.\r\n\r\n"+
			"Ваш аккаунт переведён на Free план. Созданные промпты и коллекции сохранены, но часть возможностей ограничена.\r\n\r\n"+
			"Чтобы возобновить подписку, перейдите: %s/pricing",
		planName, frontendURL,
	)
	return s.send(to, fmt.Sprintf("Подписка %s истекла — ПромтЛаб", planName), body)
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
			slog.Info("email sent", "to", to, "port", s.port)
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
