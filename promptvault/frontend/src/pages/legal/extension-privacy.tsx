import { Link } from "react-router-dom"
import { ArrowLeft } from "lucide-react"

import { AIShareBlock } from "@/components/help/ai-share-block"

export default function ExtensionPrivacyPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-3xl px-6 py-10">
        <Link
          to="/settings/integrations"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад в настройки
        </Link>

        <article className="prose-like space-y-6">
          <header className="space-y-2">
            <h1 className="text-3xl font-semibold text-foreground">
              Политика конфиденциальности Chrome-расширения
            </h1>
            <p className="text-sm text-muted-foreground">
              Расширение: <strong className="text-foreground">ПромтЛаб — библиотека AI-промптов</strong>
              <br />
              Дата вступления в силу: 11 апреля 2026 г.
              <br />
              Последнее обновление: 11 апреля 2026 г.
            </p>
          </header>

          <AIShareBlock
            mdUrl="/legal/extension-privacy.md"
            topic="разобраться в Политике конфиденциальности Chrome-расширения ПромтЛаб (что собирается, какие permissions нужны)"
            compact
          />

          <Section title="Краткая версия">
            <p>
              Расширение «ПромтЛаб» соединяется с вашим аккаунтом на promtlabs.ru через API-ключ, который вы сами
              создаёте и вставляете в расширение, и позволяет вставлять ваши промпты в поля ввода ChatGPT, Claude,
              Gemini и Perplexity. Расширение не собирает аналитику, не передаёт данные третьим лицам и не
              отправляет ничего на сторонние сервера, кроме вашего собственного backend promtlabs.ru.
            </p>
          </Section>

          <Section title="1. Какие данные собирает расширение">
            <p>
              Расширение локально хранит в <code>chrome.storage.local</code> на вашем устройстве следующую
              информацию:
            </p>
            <ul>
              <li>
                <strong>API-ключ</strong> (формат <code>pvlt_…</code>), который вы сами сгенерировали в настройках
                своего аккаунта promtlabs.ru и ввели в расширение. Ключ используется только для авторизации
                запросов к вашему backend.
              </li>
              <li>
                <strong>Адрес сервера</strong> (по умолчанию <code>https://promtlabs.ru</code>) — на какой
                экземпляр PromptVault расширение должно делать запросы.
              </li>
              <li>
                <strong>Настройка темы</strong> (светлая/тёмная/системная).
              </li>
              <li>
                <strong>Последние значения переменных промпта</strong> — чтобы при повторной вставке одного и
                того же промпта поля заполнялись автоматически. Хранятся только локально, не отправляются никуда.
              </li>
              <li>
                <strong>Последние 20 вставленных промптов</strong> — локальный кэш для вкладки «Недавние» в
                случае временной недоступности backend. Хранятся только локально.
              </li>
              <li>
                <strong>Флаг прохождения онбординга</strong> — чтобы welcome-экран не показывался повторно.
              </li>
            </ul>
            <p>
              Расширение <strong>не собирает</strong>: IP-адрес, геолокацию, историю браузера, cookies
              сторонних сайтов, содержимое страниц target-сайтов (ChatGPT, Claude, Gemini, Perplexity), любые
              персональные идентификаторы помимо API-ключа.
            </p>
          </Section>

          <Section title="2. Куда отправляются данные">
            <p>
              Все сетевые запросы расширение делает <strong>только на адрес сервера, указанный в настройках</strong>.
              По умолчанию это <code>https://promtlabs.ru</code>. Если вы настроили self-hosted сервер, запросы идут
              на ваш экземпляр.
            </p>
            <p>
              <strong>Расширение не отправляет ни одного байта ни на какие сторонние серверы</strong>: ни на
              аналитические сервисы (Google Analytics, Яндекс.Метрика, Amplitude и т. п.), ни на рекламные сети,
              ни на CDN третьих сторон. Нет трекеров, нет телеметрии, нет отчётов об ошибках на сторонние сервисы.
            </p>
            <p>
              На backend promtlabs.ru передаётся только то, что нужно для функционирования вашего аккаунта:
              заголовок <code>Authorization: Bearer pvlt_…</code>, заголовок <code>X-Client: chrome-extension/…</code>
              для аналитики использования расширения, заголовок <code>X-Timezone</code> с часовым поясом вашего
              браузера.
            </p>
          </Section>

          <Section title="3. Что расширение делает с target-сайтами (ChatGPT, Claude, Gemini, Perplexity)">
            <p>
              Для работы расширения нужны host-разрешения (<code>host_permissions</code>) на четыре сайта:
              <code> chatgpt.com</code>, <code>claude.ai</code>, <code>gemini.google.com</code>,
              <code> www.perplexity.ai</code>. Это позволяет расширению:
            </p>
            <ul>
              <li>Определять, какой из этих сайтов открыт в активной вкладке, чтобы показать соответствующий
                индикатор и разрешить вставку.</li>
              <li>Когда вы нажимаете «Вставить» — программно заполнять поле ввода сайта текстом вашего промпта.
                Это единственная операция, которую расширение выполняет на target-сайтах.</li>
            </ul>
            <p>
              Расширение <strong>не читает</strong> содержимое страниц target-сайтов, не логирует ваши переписки,
              не получает ответы AI, не отслеживает ваши действия. Оно взаимодействует только с полем ввода —
              и только по явному действию пользователя (клик «Вставить»).
            </p>
          </Section>

          <Section title="4. Permissions и зачем они нужны">
            <table className="w-full text-sm">
              <thead>
                <tr>
                  <th className="border-b border-border pb-1 text-left font-medium">Permission</th>
                  <th className="border-b border-border pb-1 text-left font-medium">Обоснование</th>
                </tr>
              </thead>
              <tbody>
                <Row
                  name="sidePanel"
                  desc="Главный UX расширения — боковая панель со списком ваших промптов."
                />
                <Row
                  name="storage"
                  desc="Локальное хранение API-ключа, настроек темы и кэша последних промптов."
                />
                <Row
                  name="activeTab"
                  desc="Доступ к активной вкладке только в момент клика пользователя — чтобы вставить текст."
                />
                <Row
                  name="scripting"
                  desc="Программная инжекция content script на target-сайтах для вставки текста в поле ввода."
                />
                <Row
                  name="host_permissions (4 сайта)"
                  desc="Ограниченный список: chatgpt.com, claude.ai, gemini.google.com, www.perplexity.ai + ваш backend. Никаких <all_urls>."
                />
              </tbody>
            </table>
          </Section>

          <Section title="5. Хранение и удаление данных">
            <p>
              Все данные расширения хранятся <strong>только локально</strong> в <code>chrome.storage.local</code>
              на вашем устройстве. На серверах PromptVault хранятся только те данные, которые вы сами
              создали через веб-интерфейс или API (промпты, коллекции, теги, команды) — это данные вашего
              аккаунта PromptVault, не данные расширения.
            </p>
            <p>
              Чтобы удалить все данные расширения, есть три способа:
            </p>
            <ol>
              <li>В расширении: «Настройки» → «Выйти». Очистит API-ключ. Остальные локальные данные сотрутся
                вместе с отзывом ключа.</li>
              <li>Отзыв API-ключа на promtlabs.ru: «Настройки» → «API-ключи» → «Удалить». Расширение получит
                401 при следующем запросе и попросит повторно ввести ключ.</li>
              <li>Удаление расширения из Chrome через <code>chrome://extensions/</code> → «Удалить». Chrome
                автоматически очистит всё локальное хранилище расширения.</li>
            </ol>
          </Section>

          <Section title="6. Передача третьим лицам">
            <p>
              <strong>Никакой передачи третьим лицам</strong>. Ни для рекламы, ни для аналитики, ни для
              исследований. Единственный получатель данных — ваш собственный экземпляр PromptVault backend.
            </p>
          </Section>

          <Section title="7. Дети и возрастные ограничения">
            <p>
              Расширение не предназначено для пользователей младше 13 лет. Мы не собираем данные детей
              намеренно. Если вы узнали, что ребёнок младше 13 лет использует расширение и связан с вашим
              аккаунтом, свяжитесь с нами для удаления данных.
            </p>
          </Section>

          <Section title="8. Изменения в политике">
            <p>
              Мы можем обновлять эту политику при добавлении новых функций в расширение. Актуальная версия
              всегда доступна по адресу <code>promtlabs.ru/legal/extension-privacy</code>. Дата последнего
              обновления указана в начале документа.
            </p>
            <p>
              Существенные изменения (добавление новых собираемых данных, новых permissions) будут анонсированы
              через обновление extension в Chrome Web Store — при установке новой версии пользователь увидит
              диалог с запросом новых прав.
            </p>
          </Section>

          <Section title="9. Контакты">
            <p>
              Если у вас есть вопросы о политике конфиденциальности или запросы на удаление данных —
              напишите нам:
            </p>
            <ul>
              <li>
                Email: <a href="mailto:slava0gpt@gmail.com" className="text-brand hover:underline">slava0gpt@gmail.com</a>
              </li>
              <li>
                GitHub:{" "}
                <a
                  href="https://github.com/slava4123/promtlab/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-brand hover:underline"
                >
                  github.com/slava4123/promtlab/issues
                </a>
              </li>
            </ul>
          </Section>

          <Section title="10. Открытый код">
            <p>
              Исходный код расширения открыт и доступен на GitHub. Вы можете самостоятельно проверить, что
              расширение делает именно то, что написано в этой политике, и ничего сверх того.
            </p>
            <p>
              <a
                href="https://github.com/slava4123/promtlab/tree/main/promptvault/extension"
                target="_blank"
                rel="noopener noreferrer"
                className="text-brand hover:underline"
              >
                github.com/slava4123/promtlab/tree/main/promptvault/extension
              </a>
            </p>
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

function Row({ name, desc }: { name: string; desc: string }) {
  return (
    <tr>
      <td className="border-b border-border/40 py-2 pr-3 font-mono text-xs text-foreground">{name}</td>
      <td className="border-b border-border/40 py-2 text-xs text-muted-foreground">{desc}</td>
    </tr>
  )
}
