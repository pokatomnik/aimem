# План: Web UI для aimem

Источник задачи: [`TASK.md`](./TASK.md).

## 1. Границы задачи

**Делаем:** React-приложение в `internal/controllers/web/ui`, сборка в `.../ui/dist`, интеграция сборки в `Makefile`, строгий TS, линт, тесты, генерация HTTP-клиента из `internal/controllers/web/openapi.yaml`.

**Не делаем (осознанно):**

- Любые правки Go-кода, `openapi.yaml` и `internal/controllers/web/server/`. `openapi.yaml` остаётся на месте. `DELETE` в спецификацию не добавляем — это сломало бы `make generate`/оген-сервер (нереализованный handler).
- Серверную раздачу `dist` — это работа по Go-части (см. §11), здесь только готовим артефакт и порядок сборки.
- Live-поиск, debounce, пагинация, тёмная тема, i18n, визуальный Markdown-редактор.

## 2. Ключевые решения

| Решение                | Выбор                                                                      | Почему                                                                                                                                                                                                                                                                                                                                                                              |
| ---------------------- | -------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Codegen                | `openapi-typescript` + `openapi-fetch`                                     | Только типы в папку, клиент 2 КБ на нативном `fetch`, `baseUrl`/`fetch` инъектируются → герметичные тесты без сети. `axios` исключён по задаче. Сгенерённые react-query-хуки (`orval`, `@hey-api/openapi-ts`) здесь не нужны: `queryFn` в наших хуках — три строки, а ключи и инвалидация остаются под нашим контролем.                                                             |
| Server state           | **TanStack Query** (`@tanstack/react-query` v5)                            | Каждый запрос имеет собственную жизнь: `loading/error/data`, отмена при смене `q`/`id`, повтор, инвалидация результатов поиска после удаления, сброс состояния мутации. Руками это два generic-хука (~80 строк), и именно там дороже всего ошибиться. Глобального клиентского стора (Redux/Zustand) не нужно: URL остаётся источником истины для `q`, форма — локальный `useState`. |
| Стили                  | CSS Modules + слой дизайн-токенов (`tokens.css`) + `global.css`            | Всё приложение — 4 экрана и ~6 примитивов. Компонентная библиотека потребовала бы кастомизации (задача прямо просит её избегать), CSS Modules дают нулевую зависимость и полный контроль над палитрой «оранжевый / чёрный / серый».                                                                                                                                                 |
| Роутер                 | `react-router` (v7), `HashRouter`                                          | Требование задачи; сервер видит только `/`. Статический сегмент `/fact/add` регистрируется рядом с `/fact/:factID` — v7 ранжирует статические выше динамических, но порядок всё равно задаём явно.                                                                                                                                                                                  |
| Компонентные примитивы | Свои `Button`/`Input`/`Spinner`/`ErrorMessage`                             | 4 штуки, ~80 строк суммарно.                                                                                                                                                                                                                                                                                                                                                        |
| Markdown               | `react-markdown` + `remark-gfm`, **без** `rehype-raw`                      | Сырой HTML не рендерим → XSS через `content` исключён by design.                                                                                                                                                                                                                                                                                                                    |
| Дата                   | Свой `formatDate` через `Intl.DateTimeFormat` с явными `locale`/`timeZone` | Детерминированные тесты, без зависимости от локали CI.                                                                                                                                                                                                                                                                                                                              |
| Тесты                  | Vitest + jsdom + Testing Library                                           | jsdom — не сеть и не диск; клиент подменяется фейком.                                                                                                                                                                                                                                                                                                                               |

### Версии пакетов

Задача требует актуальных версий, поэтому версии **не фиксируются в плане**: на шаге M0 берём последние стабильные через `npm view <pkg> version` и проверяем совместимость (`npm view <pkg> peerDependencies`, `npm ls`). Ориентир по мажорам, от которых строятся конфиги и импорты: React 19, Vite 7, `react-router` 7, `@tanstack/react-query` 5, Vitest 3, TypeScript 5.9, ESLint 9 (flat config). `package-lock.json` коммитим.

## 3. Контракт API и его подводные камни

| Экран                  | Метод и путь                   | Тело запроса                                              | Ответ                                                                      |
| ---------------------- | ------------------------------ | --------------------------------------------------------- | -------------------------------------------------------------------------- |
| `/` → `/search`        | `POST /api/v1/search`          | `{ "query": string }` (**не** `/search?q=…` — поиск POST) | `200 { "value": { "found": [{ id, createdAt, summary }] } }`               |
| `/fact/:id`            | `GET /api/v1/fact/{factID}`    | —                                                         | `200 { "value": { id, createdAt, summary, content } }`, `404` → `appError` |
| `/fact/add`            | `POST /api/v1/fact`            | `{ "value": { "fact": string } }` — двойная вложенность   | `200 { "summary": string }`                                                |
| `/fact/:id` (удаление) | `DELETE /api/v1/fact/{factID}` | —                                                         | `200 { "value": {} }` (схема `nothing`)                                    |

Что это значит для кода:

1. **`/search?q=` — только адрес состояния, не запрос.** `q` парсим через `useSearchParams`, а в сеть уходит POST-тело. Формат `?q=запрос+пользователя` обеспечивается `createSearchParams` (пробел → `+`), руками строку не клеим.
2. **`minLength: 3` у `query`** → клиентская валидация: обрезка `trim()`, кнопка заблокирована при длине < 3, подсказка «Минимум 3 символа». Это отсекает гарантированный `400`.
3. **`addFact` не возвращает ID.** После добавления перейти на страницу факта нельзя — показываем полученный `summary` и предлагаем «На главную» / «Найти похожее» (`/search?q=<summary>`). Зафиксировано в §12 как ожидаемое поведение.
4. **`appError`** `{error, message?}`. В `unwrap` превращаем в типизированный `ApiError`, коды маппим в русские тексты (§6).
5. **Удаление реализовано** операцией `removeFactByID` с ответом `{ value: {} }`. Отдельный клиент не нужен: `deleteFact` идёт через тот же сгенерированный клиент, и доменный тип не возвращается — важен только успешный статус.

## 4. Архитектура и файловая структура

Срезанный Feature-Sliced Design; зависимости строго вниз: `pages → features → entities → shared`. Никаких импортов вверх.

```
internal/controllers/web/ui/
├─ package.json                 # скрипты dev/build/test/lint/typecheck/api:generate
├─ package-lock.json
├─ index.html
├─ vite.config.ts               # base './', outDir 'dist', server.proxy /api → 127.0.0.1:8888
├─ vitest.config.ts             # jsdom, setupFiles, restoreMocks
├─ tsconfig.json                # strict-набор (§7), включает и конфиги Vite
├─ eslint.config.js             # flat config (§7)
├─ .gitignore                   # node_modules/, dist/, coverage/
├─ README.md                    # команды и ограничения контракта
└─ src/
   ├─ main.tsx                  # createRoot + <StrictMode> + QueryClientProvider
   ├─ App.tsx                   # HashRouter + <ApiProvider> + <AppRoutes>
   ├─ app/
   │  ├─ routes.tsx             # <Routes>; единственное место вызова useAppNavigation
   │  ├─ navigation/            # paths.ts (ROUTES/factPath/buildSearchPath) · useAppNavigation.ts · index.ts
   │  ├─ layout/Layout.tsx      # шапка, контейнер, подвал
   │  ├─ providers/             # ApiContext.ts · ApiProvider.tsx · useApi.ts · index.ts
   │  └─ styles/                # reset.css · tokens.css · global.css
   ├─ pages/
   │  ├─ home/        HomePage.tsx (+css) · HomePage.test.tsx
   │  ├─ search/      SearchPage.tsx (+css) · SearchResultsHeader.tsx (+css) · SearchPage.test.tsx
   │  ├─ fact/        FactPage.tsx (+css) · FactPage.test.tsx
   │  ├─ add-fact/    AddFactPage.tsx (+css)
   │  └─ not-found/   NotFoundPage.tsx (+css)
   ├─ features/
   │  ├─ search-facts/  SearchForm.tsx (+css) · SearchForm.test.tsx · useSearch.ts · useSearch.test.ts · index.ts
   │  ├─ fact-detail/   useFact.ts · index.ts
   │  ├─ add-fact/      AddFactForm.tsx (+css) · AddFactForm.test.tsx · useAddFact.ts · index.ts
   │  └─ delete-fact/   DeleteFactButton.tsx (+css) · useDeleteFact.ts · index.ts
   ├─ entities/fact/
   │  ├─ api/keys.ts            # factKeys: searchRoot / search(q) / detailRoot / detail(id)
   │  ├─ model/formatDate.ts    # (+ formatDate.test.ts)
   │  ├─ ui/                    # FactCard.tsx (+css) · MarkdownView.tsx (+css)
   │  └─ index.ts
   └─ shared/
      ├─ api/           client.ts · ApiError.ts · unwrap.ts · errorMessages.ts ·
      │                 queryClient.ts · *.test.ts · index.ts
      ├─ lib/           text.ts · paths.ts · guards.ts · *.test.ts · index.ts
      ├─ test/          renderWithProviders.tsx   # render + renderHook + фейковый Api
      └─ ui/            Button · TextInput/TextArea · Spinner · ErrorMessage · EmptyState · Meta (+ .module.css)
```

```

Инъекция зависимостей: `ApiProvider` кладёт в контекст объект `FactApi` (`search`, `getFact`, `addFact`, `deleteFact`). Страницы и хуки берут его из `useApi()`, функции слоя `entities` принимают его аргументом — так запрос остаётся чистой функцией, а `queryFn` лишь замыкает его. В тестах провайдеру передаётся фейк — реальный `fetch` не вызывается вообще. `QueryClientProvider` вкладывается снаружи `ApiProvider`; каждый тест получает собственный `QueryClient`, иначе кэш протечёт между кейсами. Роутер в тестах — `MemoryRouter` (если компоненту нужна навигация) или `HashRouter` для интеграции; обёртка собрана в `shared/test/renderWithProviders.tsx`.

## 5. Маршруты и UX

Все тексты интерфейса — русские.

### `/` — главная

- По центру: логотип-заголовок `aimem` (акцентный оранжевый), под ним — крупное поле ввода (autofocus) и кнопка с подписью «🔍» (`aria-label="Найти"`).
- Внутри `<form onSubmit>`, чтобы Enter работал так же, как кнопка. Пока идёт проверка длины < 3 — кнопка `disabled`, под полем подсказка.
- По сабмиту: `trim()` → `navigate({ pathname: '/search', search: createSearchParams({ q }).toString() })`. Никакого debounce и запросов на ввод.
- Внизу — вторичная ссылка «Добавить факт» → `/fact/add`.
- Ошибки сети здесь не возникают (запросов нет), валидация — единственный источник состояния.

### `/search?q=…`

- Шапка: компактная форма с предзаполненным `q` + кнопка «🔍», ссылки «На главную», «Добавить».
- `q` пустой/короткий → не запрашиваем API, показываем «Введите запрос» со ссылкой на `/`.
- Запрос — один на значение `q`: `useSearch(q)` с `AbortController` и защитой от повторного запуска (в dev `StrictMode` монтирует эффект дважды — см. §12). Переход назад/вперёд переиспользует URL как единственное состояние.
- Состояния: скелетон загрузки → список результатов → «Ничего не найдено» → блок ошибки с кнопкой «Повторить» (`aria-live="polite"` на контейнере результатов). Скелетон — только на `isPending`; при `isFetching` на уже показанном списке (возврат по истории, повтор) держим старые результаты приглушёнными и не мигаем пустым экраном.
- Элемент результата: `summary` (основной текст, ~1.05rem), под ним моноширинным серым — `id` и `createdAt`. Вся карточка — `<Link to={'/fact/' + id}>`. Учитываем требование «только ID + summary»: `content` не подтягиваем.
- Шапка результатов: «Найдено: N».

### `/fact/:factID`

- `GET`. Заголовок — `summary`, под ним `createdAt` серым, тело — `content` через `MarkdownView`.
- Кнопка «← Назад» → `/` (по требованию задачи), визуально вторичная. Ссылки внутри Markdown — `target="_blank"` + `rel="noopener noreferrer"`.
- Кнопка «Удалить» — акцентно-контурная, реализует двухшаговое подтверждение inline («Удалить?» / «Да, удалить» / «Отмена») вместо `window.confirm`: без модалок, чисто тестируется.
- После успешного удаления — переход на `/`. Ошибка — сообщение из `ApiError` с сохранением кнопок.
- `404`/`FACT_DOES_NOT_EXIST` → отдельное состояние «Факт не найден» + «На главную» (не общий блок ошибки).

### `/fact/add`

- `<textarea>` с плейсхолдером-примером, валидация — непустая строка после `trim()`; счётчик символов опционально.
- Сабмит: `POST /api/v1/fact` с `{ value: { fact } }`. Кнопка в состоянии загрузки заблокирована (двойной сабмит невозможен).
- Успех: карточка с «Факт добавлен» и возвращённым `summary` + действия «Добавить ещё» (сброс формы), «На главную», «Найти похожее» (`/search?q=<summary>`).
- Ошибка: текст по коду (`ALREADY_HAS_FACT`, `SUMMARIZE`, `GET_EMBEDDINGS`, …) + «Повторить». Введённый текст сохраняется.

## 6. Слой API

- `vite.config.ts`: `base: './'` (ассеты относительные — работает и при раздаче из подкаталога), `build.outDir: 'dist'`, `emptyOutDir: true`, `sourcemap: false`.
- `src/shared/api/client.ts`: `createFactApi({ baseUrl?, fetchImpl? })`, внутри — `createClient<paths>(…)`. В рантайме `baseUrl` не передаётся вообще: запросы идут на `/api/v1/...` относительно текущего хоста. Параметр существует только как тестовый шов — `Request` в Node не принимает относительные URL.
- `dev`-режим: `server.proxy = { '/api': 'http://127.0.0.1:8888' }`, чтобы `npm run dev` работал против живого бекенда без CORS и без правок Go.
- `unwrap.ts`: `{ data, error, response }` → либо `data`, либо `throw new ApiError(code, message, status)`. Код берётся из `error`, при отсутствии тела — `UNKNOWN` с HTTP-статусом; сетевой сбой fetch — `NETWORK_ERROR`.
- `errorMessages.ts`: `Record<ErrorCode | LocalErrorCode, string>`, exhaustive `switch`/`satisfies` (компилятор ломается при добавлении нового кода в спецификацию — это желаемое поведение):

  | Код                                                                                       | Текст                                                        |
  | ----------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
  | `ALREADY_HAS_FACT`                                                                        | Такой факт уже есть                                          |
  | `CANNOT_GET_MODELS`, `GET_EMBEDDINGS`, `SUMMARIZE`                                        | Не удалось обратиться к Ollama — проверьте, что она запущена |
  | `EMPTY_EMBEDDINGS`, `EMBEDDINGS_QUANTITY_NOT_EQUAL_TO_FACTS_QUANTITY`, `DB_CONVERT_ERROR` | Ошибка хранилища. Попробуйте ещё раз                         |
  | `FACT_DOES_NOT_EXIST`, `NOT_FOUND`                                                        | Факт не найден                                               |
  | `BAD_REQUEST`                                                                             | Некорректный запрос                                          |
  | `INTERNAL_ERROR`, `UNKNOWN`                                                               | Что-то пошло не так                                          |
  | `NETWORK_ERROR`                                                                           | Сервер недоступен                                            |

- `entities/fact/api/*` — по функции на эндпоинт, каждая возвращает доменный тип (`Fact`, `FactListItem`), а не `{value: …}`. Это единственное место, где известна вложенность ответов.

### Политика кэша и повторов

```ts
// shared/api/queryClient.ts
export const createQueryClient = (): QueryClient =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
        staleTime: 30_000,
        gcTime: 5 * 60_000,
      },
    },
  });
```

- **`retry: false`** — дефолтные три повтора умножат один неудачный поиск на четыре POST-запроса к бекенду, каждый из которых дергает эмбеддинги и модель суммаризации. Ошибка доходит до пользователя сразу, с кнопкой «Повторить». `useMutation` по умолчанию не ретраится — для `POST /fact` это то, что нужно.
- **`staleTime: 30s`** — сценарий «поиск → факт → назад» отдаёт те же результаты из кэша вместо скелетона. Без этого каждый возврат по истории выглядит как новый поиск.
- **`refetchOnWindowFocus: false`** — возврат фокуса в окно не должен молча перезапускать поиск.
- Ключи — рядом с API в `entities/fact/api/keys.ts`: `factKeys.detail(id)`, `factKeys.search(q)`, `factKeys.searchRoot` для инвалидации. В ключ попадает `trim()`-нормализованный `q`, иначе `' foo'` и `'foo'` создадут две записи кэша на один результат.
- После добавления — `invalidateQueries({ queryKey: factKeys.searchRoot })`; после удаления — дополнительно `removeQueries({ queryKey: factKeys.detail(id) })`.
- `signal` из `QueryFunctionContext` прокидывается в `openapi-fetch`: `client.POST('/api/v1/search', { body, signal })`. При смене `q` запрос действительно отменяется, а не просто игнорируется.

## 7. TypeScript и линт

`tsconfig.json`: `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride`, `noPropertyAccessFromIndexSignature`, `noUnusedLocals`, `noUnusedParameters`, `noFallthroughCasesInSwitch`, `verbatimModuleSyntax`, `isolatedModules`, `useDefineForClassFields`, `jsx: react-jsx`. Отдельный `tsconfig.node.json` для `vite.config.ts`/`eslint.config.js`.

ESLint 9 flat config: `@eslint/js`, `typescript-eslint` в пресете `strictTypeChecked` + `stylisticTypeChecked`, `eslint-plugin-react`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`.

Обязательные правила (задача требует максимально строгую типизацию):

- `@typescript-eslint/no-explicit-any: error`;
- `no-unsafe-assignment`/`-member-access`/`-call`/`-return`/`-argument` (уже в `strictTypeChecked`) — `error`;
- `@typescript-eslint/switch-exhaustiveness-check: error`;
- `@typescript-eslint/consistent-type-imports: error`;
- `@typescript-eslint/no-floating-promises: error`, `no-misused-promises: error`;
- `react-hooks/rules-of-hooks: error`, `react-hooks/exhaustive-deps: error`;
- `--max-warnings=0` в скрипте `lint`.

`src/api/generated/**` — в `ignores` ESLint (код сгенерирован), но **не** в `exclude` tsconfig: типы должны проверяться.

Скрипты `package.json`:

```json
{
  "scripts": {
    "dev": "vite",
    "api:generate": "openapi-typescript ../openapi.yaml -o src/api/generated/schema.d.ts",
    "build": "tsc --noEmit && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint . --max-warnings=0",
    "typecheck": "tsc --noEmit"
  }
}
```

Путь `../openapi.yaml` — из `internal/controllers/web/ui` в `internal/controllers/web/openapi.yaml`; спецификация не перемещается. Сгенерированный файл коммитим (сборка не требует codegen-шага и не зависит от версии инструмента).

## 8. Дизайн

Токены в `tokens.css` (единственное место с цветами):

```
--color-accent:        #F97316;
--color-accent-strong: #EA580C;   /* hover/active */
--color-accent-soft:   #FFF3E8;   /* подсветка, focus-ring фон */
--color-text:          #0B0B0C;   /* весь основной текст */
--color-text-secondary:#6B7280;   /* secondary */
--color-bg:            #FFFFFF;
--color-surface:       #FAFAFA;
--color-border:        #E5E7EB;
--radius-*  --space-*  --shadow-*  --font-sans: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif
```

Правила вёрстки:

- Никаких внешних ассетов: ни Google Fonts, ни CDN-иконок — всё должно лежать в `dist` (иначе UI зависит от сети, а Go-бинарь должен оставаться самодостаточным). Единственная «иконка» — emoji 🔍 в кнопке, как просили.
- Ширина контента: главная — центрированный блок `max-width: 640px`; результаты — `max-width: 720px`, левое выравнивание.
- `:focus-visible` — 2px кольцо `--color-accent` с отступом; контраст текста ≥ 4.5:1 (серый взят с запасом).
- Переходы — `transition` 120–160 мс, отключаются через `@media (prefers-reduced-motion: reduce)`.
- Пустые состояния и скелетоны (серые блоки с `--color-border`) вместо «пустого экрана».
- Мобильная адаптация: `padding` и размер заголовка через `clamp()`, кнопки ≥ 44px по высоте.

## 9. Тесты

Правила из `AGENTS.md` переносятся буквально: **никакого I/O в тестах**. Для фронтенда это значит: без сети (никакого MSW на реальном сокете, никаких `httptest`-аналогов), без `localStorage`/`sessionStorage`, без чтения файлов и без зависимости от системной локали и текущего времени. `fetch` в тестах либо не вызывается (фейковый `Api` в провайдере), либо подменён функцией, возвращающей `new Response(...)` — это не сетевой вызов.

Покрытие — прежде всего бизнес-логика:

| Файл                                     | Что проверяем                                                                                                                                                                                                                                                                          |
| ---------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `shared/lib/text.test.ts`                | `trim`, минимум 3 символа для поиска, минимум 1 для факта, границы (2/3, пустая строка, пробелы)                                                                                                                                                                                       |
| `app/routes.test.ts`                     | `buildSearchPath` даёт `#/search?q=a+b` (пробел → `+`), кириллица кодируется, спецсимволы `&`, `#` не ломают URL                                                                                                                                                                       |
| `shared/api/unwrap.test.ts`              | успех → `data`; `appError` → `ApiError` с кодом и статусом; пустое тело → `UNKNOWN`; сбой `fetch` → `NETWORK_ERROR`                                                                                                                                                                    |
| `shared/api/errorMessages.test.ts`       | каждый код из спецификации имеет текст; неизвестный код → дефолт                                                                                                                                                                                                                       |
| `entities/fact/model/formatDate.test.ts` | фиксированная дата + фиксированные `locale`/`timeZone` → стабильная строка; невалидная дата → `—`                                                                                                                                                                                      |
| `entities/fact/api/*.test.ts`            | двойная вложенность `{value: {fact}}` / `{value: …}` и распаковка ответа; для `deleteFact` — что ушёл `DELETE` на `/api/v1/fact/{id}` (фейковый `fetch`, аргументы проверяются `expect`)                                                                                               |
| `features/*/*.test.ts(x)`                | `renderHook` со свежим `QueryClient` (`retry: false`, `gcTime: 0`) и фейковым `Api`: `pending → success`, `pending → error`, ошибка не ретраится, повторный рендер с тем же ключом не шлёт второй запрос, мутация после `reset()` снова `idle`, смена `q` шлёт ровно один новый запрос |
| `pages/*.test.tsx`                       | дом → сабмит ведёт на `/search?q=…`; поиск рендерит `summary`+`id`; факт рендерит Markdown-заголовок и работает «Назад»; add-форма шлёт текст и показывает `summary`; удаление при ошибке показывает сообщение                                                                         |

`vitest.config.ts`: `environment: 'jsdom'`, `setupFiles: ['src/test/setup.ts']`, `restoreMocks: true`, `clearMocks: true`. В setup — только `@testing-library/jest-dom` и заглушки несуществующих в jsdom API (`matchMedia`), никакого I/O. `test` не должен требовать сети: прогон в офлайне — обязательная проверка перед сдачей.

Сквозных (Playwright) тестов не делаем: они поднимают сервер и браузер, что прямо противоречит требованию к тестам.

## 10. Интеграция в Makefile

Добавляемые цели (Go-код не трогаем; `build` теперь сначала собирает UI):

```make
NPM      ?= npm
UI_DIR   := internal/controllers/web/ui
UI_DIST  := $(UI_DIR)/dist
UI_STAMP := $(UI_DIR)/node_modules/.install-stamp

.PHONY: ... ui-install ui-build ui-test ui-lint ui-typecheck

$(UI_STAMP): $(UI_DIR)/package.json $(UI_DIR)/package-lock.json
	cd $(UI_DIR) && $(NPM) ci
	@touch $(UI_STAMP)

ui-install: $(UI_STAMP)
ui-build: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run build
ui-typecheck: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run typecheck
ui-test: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run test
ui-lint: $(UI_STAMP)
	cd $(UI_DIR) && $(NPM) run lint

build: ui-build
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build ...

clean:
	rm -rf $(BUILD_DIR) $(UI_DIST)
```

`node_modules/.install-stamp` пересобирает зависимости только при изменении `package.json`/`package-lock.json`, поэтому `make build` не качает пакеты каждый раз. `make test` остаётся Go-only, чтобы не замедлять основной цикл; фронтовые проверки — `make ui-test` и `make ui-lint` (обе гоняются перед PR).

Готовность: `make clean && make build` с нуля создаёт `internal/controllers/web/ui/dist/index.html` + `dist/assets/*`, затем собирает Go-бинарь; `make ui-test` и `make ui-lint` зелёные.

## 11. Что останется на бекенде (вне этой задачи)

Чтобы `dist` отдавался, нужен `//go:embed all:dist` в пакете `internal/controllers/web/ui`
и обработчик, возвращающий `index.html` на `/`. SPA-фолбэк не нужен: `HashRouter`
держит всё после `#`, сервер видит только `/`. `make build` уже гарантирует, что
`dist` существует до компиляции Go, иначе `go:embed` упал бы на пустом каталоге.

Это единственный оставшийся кусок, и он вне scope задачи (Go-код не трогаем).

## 12. Принятые решения

Вопросы ниже закрыты в обсуждении и считаются частью постановки: отступление от них — повод пересмотреть задачу, а не «риск».

1. **`dist` не коммитим.** Каталог сборки и `node_modules/` живут в `ui/.gitignore`. Артефакт получается только через `make build` (или `npm ci && npm run build` в CI, если он собирает Go напрямую).
2. **`addFact` не возвращает `id` — так и должно быть.** Переход на страницу факта после добавления не нужен и не планируется. Экран успеха показывает `summary` и действия «Добавить ещё» / «На главную» / «Найти похожее». Кода под гипотетический `id` не пишем.
3. **`DELETE` реализован** (`removeFactByID`, ответ `{ value: {} }`). Кнопка использует тот же сгенерированный клиент, что и остальные операции; отдельного типа-обходного пути не осталось.
4. **`StrictMode` в dev — не проблема.** В тестах реальных запросов нет (`Api` подменяется фейком), а в рантайме дублирование гасит дедупликация TanStack Query по ключу. Отдельных мер не предпринимаем.
5. **Сгенерированный код исключён из ESLint** (`src/api/generated/**` в `ignores`), но проверяется `tsc` вместе с остальным кодом.
6. **Точные версии пакетов в документации не фиксируются.** На M0 берутся последние стабильные, единственный источник истины о версиях — `package-lock.json`.
7. **Всё UI-тексты — на английском.** Ни интернационализации, ни русского: копирайт пишется прямо в компонентах.
8. **Абсолютные импорты через `@/`.** `tsconfig.json` (`baseUrl` + `paths`) и `resolve.alias` в `vite.config.ts`; `vitest.config.ts` мержит конфиг Vite, чтобы алиас имел один источник истины.
9. **`useNavigate` — только в `app`.** Все инварианты навигации собраны в `app/navigation/useAppNavigation.ts` (`AppNavigation`), страницы получают типизированные колбэки и href'ы через `props`. `paths.ts` живёт там же, в `shared/lib` его нет.
10. **`baseURL` не хардкодится.** Клиент всегда вызывает `/api/v1/...` относительно текущего хоста: UI и API отдаёт один и тот же бинарь, разделения на домены не будет. `baseUrl` остался только как тестовый шов (Node не резолвит относительные URL в `Request`).
11. **Порог `minLength: 3` зашит навсегда.** Бекенд может только ослабить ограничение и никогда не ужесточит его ниже трёх символов, поэтому клиент не подстраивается: `SEARCH_QUERY_MIN_LENGTH = 3` — фиксированная константа в `shared/lib/text.ts`, не читается из конфигурации.

## 13. Порядок работ и статус

| #   | Шаг | Статус |
| --- | --- | --- |
| M0  | Скаффолд: `package.json`, `tsconfig.json`, `eslint.config.js`, `vite.config.ts`, `vitest.config.ts`, `index.html`, `src/main.tsx`, `.gitignore` | **готово** |
| M1  | Codegen и слой API: `api:generate`, `client.ts`, `queryClient.ts`, `ApiProvider`/`useApi`, `unwrap`, `ApiError`, `errorMessages`, `renderWithProviders` | **готово** |
| M2  | Дизайн-система: `reset.css`, `tokens.css`, `global.css`, `Button`/`TextInput`/`TextArea`/`Spinner`/`ErrorMessage`/`Meta`, `Layout` | **готово** |
| M3  | Роутер и главная: `App.tsx`, `routes.tsx`, `paths.ts` (+тест), `HomePage`, `SearchForm` (+тест) | **готово** |
| M4  | Поиск: `factKeys`, `useSearch` (+тест), `SearchPage` (+тест), `FactCard` | **готово** |
| M5  | Факт: `useFact`, `MarkdownView`, `FactPage` (+тест), «Назад», `useDeleteFact` + `DeleteFactButton` | **готово** |
| M6  | Добавление: `useAddFact`, `AddFactForm` (+тест), `AddFactPage` | **готово** |
| M7  | Makefile: `ui-install`/`ui-build`/`ui-test`/`ui-lint`/`ui-typecheck`, `build: ui-build`, `clean`; `README.md` | **готово** |
| M8  | Финальная проверка: `make clean && make build`, `make ui-test`, `make ui-lint`, `make test`, `go vet ./...` | **готово** |
| M9  | Правки по ревью: английские тексты, алиас `@/`, навигация в `app` (`useAppNavigation`), `EmptyState`, `SearchResultsHeader`, относительный `baseURL`, DELETE через сгенерированный клиент | **готово** |

### Результат проверки

- `make clean && make build` — собирает `dist/` (index.html + css + js) и затем Go-бинарь `build/aimem`.
- `make ui-test` — 78 тестов в 12 файлах, все зелёные; прогон не требует сети.
- `make ui-lint` — 0 ошибок, 0 предупреждений (`--max-warnings=0`).
- `npm run typecheck` — чисто (strict + `noUncheckedIndexedAccess` + `exactOptionalPropertyTypes`).
- `make test` и `go vet ./...` — Go-часть не сломана.

## 14. Соответствие требованиям TASK.md

| Требование                                                           | Где закрыто                      |
| -------------------------------------------------------------------- | -------------------------------- |
| React + TypeScript + Vite + Vitest                                   | §2, §13/M0                       |
| `npm run build` и встраивание в Makefile                             | §10, §13/M7                      |
| Исходники в `internal/controllers/web/ui`, сборка в `.../ui/dist`    | §4, `vite.config.ts`             |
| Учёт `embed` при раздаче `dist`                                      | §10 (порядок сборки), §11, §12.1 |
| Go-файлы не трогаются                                                | §1                               |
| Оранжевый / чёрный / серый, аккуратная вёрстка                       | §8                               |
| `HashRouter`, пути `/`, `/search?q=`, `/fact/{id}`, `/fact/add`      | §2, §5                           |
| Главная: центр, кнопка «🔍», переход на `/search`, без live-поиска   | §5                               |
| Поиск: результаты = ID + summary, поле сверху                        | §5                               |
| Факт: Markdown, «Назад» на главную, «Удалить» через HTTP-клиент      | §5, §3.5                         |
| Добавление: простой текстовый редактор                               | §5                               |
| Клиент из `openapi.yaml`, скрипт в `package.json`, транспорт `fetch` | §3, §7, `api:generate`           |
| Разделение логики и UI, решение по стейт-менеджеру                   | §2, §4, §6 (политика кэша)       |
| Тесты бизнес-логики без побочных эффектов                            | §9                               |
| `npm run test` + `npm run lint`, запрет `any`, строгий TS            | §7, §9                           |
| Архитектура: FSD, DRY/SOLID/KISS                                     | §4                               |
