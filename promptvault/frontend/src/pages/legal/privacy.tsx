import { Link } from "react-router-dom"
import { ArrowLeft } from "lucide-react"

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-3xl px-6 py-10">
        <Link
          to="/"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          На главную
        </Link>

        <article className="prose-like space-y-6">
          <header className="space-y-2">
            <h1 className="text-3xl font-semibold text-foreground">
              Политика конфиденциальности
            </h1>
            <p className="text-sm text-muted-foreground">
              Сервис: <strong className="text-foreground">ПромтЛаб (promtlabs.ru)</strong>
              <br />
              Дата вступления в силу: 12 апреля 2026 г.
              <br />
              Последнее обновление: 12 апреля 2026 г.
            </p>
          </header>

          <Section title="1. Какие данные мы собираем">
            <p>При использовании Сервиса мы собираем:</p>
            <ul>
              <li>
                <strong>Данные аккаунта</strong> — email, имя, имя пользователя, аватар (при загрузке).
                Пароль хранится в виде bcrypt-хэша, недоступного для чтения.
              </li>
              <li>
                <strong>OAuth-данные</strong> — при входе через GitHub, Google или Яндекс мы получаем
                только идентификатор и email от провайдера. Мы не получаем доступ к вашим репозиториям,
                файлам или контактам.
              </li>
              <li>
                <strong>Пользовательский контент</strong> — промпты, коллекции, теги, комментарии к версиям.
                Этот контент принадлежит вам.
              </li>
              <li>
                <strong>Данные использования</strong> — количество использований промпта, дата последнего
                доступа, активность для расчёта стриков. Используются для аналитики и бейджей.
              </li>
              <li>
                <strong>Технические данные</strong> — IP-адрес (для rate limiting), User-Agent (для
                аудита), часовой пояс (для корректного отображения стриков).
              </li>
            </ul>
          </Section>

          <Section title="2. Как мы используем данные">
            <ul>
              <li>Предоставление и поддержка функций Сервиса</li>
              <li>Аутентификация и защита аккаунта</li>
              <li>Отправка служебных email (верификация, сброс пароля)</li>
              <li>Расчёт лимитов тарифного плана</li>
              <li>Улучшение Сервиса на основе агрегированной статистики</li>
            </ul>
            <p>
              Мы <strong>не используем</strong> ваши данные для рекламы, профилирования
              или продажи третьим лицам.
            </p>
          </Section>

          <Section title="3. AI-обработка данных">
            <p>
              Функции «Улучшить», «Переписать», «Анализ» и «Вариации» отправляют содержимое промпта
              на сервер OpenRouter (API-провайдер) для обработки моделью Claude. Данные обрабатываются
              в соответствии с{" "}
              <a
                href="https://openrouter.ai/privacy"
                target="_blank"
                rel="noopener noreferrer"
                className="text-brand hover:underline"
              >
                политикой конфиденциальности OpenRouter
              </a>
              . Мы не храним результаты AI-обработки на своих серверах — они передаются напрямую
              в ваш браузер через SSE-стриминг.
            </p>
          </Section>

          <Section title="4. Хранение данных">
            <p>
              Данные хранятся на управляемом сервере PostgreSQL в дата-центре Timeweb Cloud
              (Россия, Санкт-Петербург), в соответствии с требованиями ФЗ-152 о персональных данных.
            </p>
            <ul>
              <li>Удалённые промпты хранятся в корзине 30 дней, после чего удаляются безвозвратно</li>
              <li>Версии промптов хранятся бессрочно для возможности отката</li>
              <li>Логи аудита (для администраторов) хранятся бессрочно</li>
              <li>Коды верификации email удаляются через 15 минут</li>
            </ul>
          </Section>

          <Section title="5. Передача третьим лицам">
            <p>Мы передаём данные третьим лицам только в следующих случаях:</p>
            <ul>
              <li>
                <strong>OpenRouter</strong> — содержимое промпта при использовании AI-функций
                (только по вашему явному действию)
              </li>
              <li>
                <strong>SMTP-провайдер</strong> — email-адрес для отправки служебных писем
              </li>
              <li>
                <strong>По требованию закона</strong> — при наличии законного запроса
                уполномоченных органов РФ
              </li>
            </ul>
            <p>
              Мы <strong>не продаём</strong> и <strong>не передаём</strong> данные рекламным
              сетям, аналитическим сервисам или иным коммерческим третьим лицам.
            </p>
          </Section>

          <Section title="6. Безопасность">
            <p>Мы применяем следующие меры защиты:</p>
            <ul>
              <li>HTTPS (TLS) для всех соединений</li>
              <li>Bcrypt-хэширование паролей</li>
              <li>JWT-токены с коротким временем жизни (15 минут)</li>
              <li>HttpOnly cookies для refresh-токенов</li>
              <li>Rate limiting для защиты от brute force</li>
              <li>TOTP 2FA для административных действий</li>
              <li>Аудит-лог всех административных операций</li>
            </ul>
          </Section>

          <Section title="7. Ваши права">
            <p>Вы имеете право:</p>
            <ul>
              <li><strong>Исправление</strong> — обновить данные профиля в настройках</li>
              <li><strong>Удаление</strong> — удалить аккаунт и все связанные данные</li>
              <li><strong>Отзыв согласия</strong> — отключить OAuth-привязки, удалить API-ключи</li>
            </ul>
            <p>
              Для реализации этих прав обратитесь по email{" "}
              <a href="mailto:slava0gpt@gmail.com" className="text-brand hover:underline">
                slava0gpt@gmail.com
              </a>
              .
            </p>
          </Section>

          <Section title="8. Cookies">
            <p>
              Сервис использует только технически необходимые cookies:
            </p>
            <ul>
              <li><strong>Refresh token</strong> — HttpOnly cookie для автоматического продления сессии</li>
              <li><strong>OAuth state</strong> — временная cookie для безопасности OAuth-авторизации</li>
            </ul>
            <p>
              Мы не используем аналитические, рекламные или трекинговые cookies.
            </p>
          </Section>

          <Section title="9. Изменения в политике">
            <p>
              Мы можем обновлять эту политику. Актуальная версия всегда доступна на данной странице.
              О существенных изменениях мы уведомим через email или уведомление в Сервисе.
            </p>
          </Section>

          <Section title="10. Контакты">
            <p>По вопросам конфиденциальности и защиты данных:</p>
            <ul>
              <li>
                Email:{" "}
                <a href="mailto:slava0gpt@gmail.com" className="text-brand hover:underline">
                  slava0gpt@gmail.com
                </a>
              </li>
            </ul>
          </Section>
        </article>

        <footer className="mt-10 border-t border-border pt-6 text-xs text-muted-foreground">
          © 2026 ПромтЛаб. Все права защищены.
        </footer>
      </div>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="text-lg font-semibold text-foreground">{title}</h2>
      <div className="space-y-3 text-sm leading-relaxed text-muted-foreground">{children}</div>
    </section>
  )
}
