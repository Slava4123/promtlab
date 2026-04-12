// Package main — CLI для промоута существующего пользователя до admin +
// опционального первичного TOTP enrollment.
//
// Используется при bootstrap (создание первого админа) и для ручного promote
// support-инженеров. Работает локально на сервере (не через HTTP) — требует
// прямой доступ к БД и .env файлу.
//
// Использование:
//
//	go run ./cmd/create-admin --email=admin@example.com
//	./create-admin --email=admin@example.com --enroll-totp=false
//	./create-admin --email=admin@example.com --skip-confirm
//
// Flow:
//  1. Load config через config.Load() — читает .env, как и сервер.
//  2. Connect к PostgreSQL.
//  3. Найти юзера по email. Если не найден — exit 1.
//  4. Если уже role=admin — warning, exit 0 (идемпотентно).
//  5. UPDATE users SET role='admin'.
//  6. (опционально) Запустить TOTP enrollment через adminauthuc.Service.Enroll:
//     вывести secret + QR URL + 10 backup codes со scary warning.
//
// Backup codes показываются **только один раз** — как в GitHub/Google.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"promptvault/internal/infrastructure/config"
	"promptvault/internal/infrastructure/postgres"
	pgrepo "promptvault/internal/infrastructure/postgres/repository"
	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
	adminauthuc "promptvault/internal/usecases/adminauth"
)

func main() {
	var (
		email       = flag.String("email", "", "email существующего пользователя (обязательный)")
		enrollTOTP  = flag.Bool("enroll-totp", true, "сразу запустить TOTP enrollment")
		skipConfirm = flag.Bool("skip-confirm", false, "не ждать подтверждения (для scripted use)")
	)
	flag.Parse()

	if *email == "" {
		fmt.Fprintln(os.Stderr, "error: --email обязателен")
		flag.Usage()
		os.Exit(1)
	}

	// slog в text-формате — читаемо в терминале.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: не удалось загрузить конфиг: %v\n", err)
		os.Exit(1)
	}

	db, err := postgres.Connect(cfg.Database, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: не удалось подключиться к БД: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	userRepo := pgrepo.NewUserRepository(db)

	user, err := userRepo.GetByEmail(ctx, *email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "error: пользователь с email %q не найден\n", *email)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: не удалось найти пользователя: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Пользователь найден: id=%d, email=%s, name=%q\n", user.ID, user.Email, user.Name)

	if user.Role == models.RoleAdmin {
		fmt.Printf("⚠ Пользователь уже имеет role=admin — пропускаю обновление\n")
	} else {
		if !*skipConfirm {
			fmt.Printf("\nПромоут %s (id=%d) до admin? [y/N] ", user.Email, user.ID)
			if !readYes() {
				fmt.Println("Отменено.")
				os.Exit(0)
			}
		}

		user.Role = models.RoleAdmin
		if err := userRepo.Update(ctx, user); err != nil {
			fmt.Fprintf(os.Stderr, "error: не удалось обновить роль: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Роль обновлена на 'admin'\n")
	}

	if !*enrollTOTP {
		fmt.Println("\nTOTP enrollment пропущен (--enroll-totp=false).")
		fmt.Println("Админ сможет запустить enrollment через UI при первом login.")
		return
	}

	// TOTP enrollment.
	totpRepo := pgrepo.NewTOTPRepository(db)
	adminauthSvc := adminauthuc.NewService(totpRepo, userRepo)

	result, err := adminauthSvc.Enroll(ctx, user.ID)
	if err != nil {
		if errors.Is(err, adminauthuc.ErrTOTPAlreadyConfirmed) {
			fmt.Println("\n⚠ TOTP уже настроен и подтверждён для этого пользователя.")
			fmt.Println("Если нужно сбросить, используйте UI (admin → Disable TOTP).")
			return
		}
		fmt.Fprintf(os.Stderr, "error: TOTP enrollment failed: %v\n", err)
		os.Exit(1)
	}

	printEnrollResult(user.Email, result)
}

func printEnrollResult(email string, r *adminauthuc.EnrollResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("TOTP ENROLLMENT — сохраните данные ниже прямо сейчас!")
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println()
	fmt.Printf("Пользователь:  %s\n", email)
	fmt.Printf("TOTP secret:   %s\n", r.Secret)
	fmt.Printf("QR URL:        %s\n", r.QRURL)
	fmt.Println()
	fmt.Println("Отсканируйте QR URL в Authenticator (Google / 1Password / Authy) или")
	fmt.Println("введите secret вручную. Тип: Time-based, 6 цифр, SHA1, интервал 30с.")
	fmt.Println()
	fmt.Println("⚠ BACKUP CODES — показываются ОДИН РАЗ. Сохраните их в password manager:")
	fmt.Println()
	for i, code := range r.BackupCodes {
		fmt.Printf("   %2d.  %s\n", i+1, code)
	}
	fmt.Println()
	fmt.Println("Каждый код можно использовать только один раз. Если потеряли телефон —")
	fmt.Println("логиньтесь через backup code и сразу regenerate через /admin/totp.")
	fmt.Println()
	fmt.Println("Далее: войдите в UI обычным логином, затем введите текущий TOTP код")
	fmt.Println("из Authenticator для confirm enrollment. До confirm логин админа работает")
	fmt.Println("без TOTP (чтобы не заблокировать доступ).")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
}

// readYes читает stdin и возвращает true если пользователь ввёл y/yes.
func readYes() bool {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}
