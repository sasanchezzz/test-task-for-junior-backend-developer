# Test Task for Junior Backend Developer

HTTP-сервис для управления задачами на Go. Проект реализует CRUD для задач, поддерживает повторяющиеся задачи, рассчитывает ближайшие даты выполнения и позволяет сохранить выбранные даты в отдельную таблицу scheduled tasks.

## Стек

- Go 1.23
- PostgreSQL 16
- Gorilla Mux
- pgx
- Docker и Docker Compose
- Swagger UI / OpenAPI

## Возможности

- Создание, получение, обновление, удаление и список задач.
- Статусы задач: `new`, `in_progress`, `done`.
- Повторяемость задач через поле `recurrence`.
- Расчет ближайших дат выполнения задачи.
- Проверка дат на выходные и праздничные дни.
- Возврат альтернативных рабочих дат, если дата попадает на нерабочий день.
- Сохранение выбранных дат выполнения в таблицу `scheduled_tasks`.

## Пример использования

Пользователь:
  - Создает задачу, в которой можно указать определенную периодичность в поле `recurrence`, детали в разделе "Типы повторяемости".
  - Внутри `recurrence` можно задать `start_date` и `end_date`, чтобы ограничить период выполнения задачи.
  - Запрашивает список ближайших дат выполнения задачи по её ID.
  - Если часть дат попадает на выходной или праздник, сервис возвращает предупреждение и предлагает альтернативные рабочие даты, детали в разделе "Примеры запросов: Рассчитать ближайшие даты выполнения".
  - Выбирает подходящие даты: оставляет исходные или заменяет их предложенными альтернативами.

Сервис:
  - Рассчитывает ближайшие даты выполнения задачи на основе типа повторяемости.
  - Проверяет даты на выходные и праздничные дни.
  - Возвращает предупреждения и альтернативные рабочие даты для нерабочих дней, детали в разделе "Нерабочие дни"
  - Сохраняет, выбранный пользователем, список дат в таблице scheduled_tasks: id записи, id задачи, статус задачи, дата задачи.

Для этого используются маршруты:

- `POST /api/v1/tasks` - создать задачу.
- `GET /api/v1/tasks/{id}/upcoming-dates?count=10` - рассчитать ближайшие даты.
- `POST /api/v1/tasks/{id}/upcoming-dates` - сохранить выбранные даты.

## Быстрый запуск

Из корня проекта:

```bash
docker compose up --build
```

После запуска:

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/openapi.json`
- PostgreSQL с хоста: `localhost:5435`

Если база уже запускалась раньше со старой схемой, пересоздайте volume:

```bash
docker compose down -v
docker compose up --build
```

Это нужно потому, что SQL-файлы из `migrations` монтируются в `/docker-entrypoint-initdb.d` и выполняются PostgreSQL только при первой инициализации пустого volume.

## Переменные окружения

Приложение использует следующие переменные:

| Переменная | Значение по умолчанию | Описание |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Адрес HTTP-сервера |
| `DATABASE_DSN` | `postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable` | DSN подключения к PostgreSQL |

В `docker-compose.yml` для контейнера приложения используется DSN:

```text
postgres://postgres:postgres@postgres:5432/taskservice?sslmode=disable
```

## API

Базовый префикс:

```text
/api/v1
```

Маршруты:

| Метод | Путь | Описание |
| --- | --- | --- |
| `POST` | `/api/v1/tasks` | Создать задачу |
| `GET` | `/api/v1/tasks` | Получить список задач |
| `GET` | `/api/v1/tasks/{id}` | Получить задачу по ID |
| `PUT` | `/api/v1/tasks/{id}` | Обновить задачу |
| `DELETE` | `/api/v1/tasks/{id}` | Удалить задачу |
| `GET` | `/api/v1/tasks/{id}/upcoming-dates?count=10` | Рассчитать ближайшие даты выполнения |
| `POST` | `/api/v1/tasks/{id}/upcoming-dates` | Сохранить выбранные даты выполнения |

## Модель задачи

Пример тела запроса для создания или обновления задачи:

```json
{
  "title": "Подготовить отчет",
  "description": "Собрать данные и отправить отчет",
  "status": "new",
  "recurrence": {
    "type": "every_n_days",
    "days_count": 3,
    "start_date": "2026-04-21T09:00:00Z",
    "end_date": "2026-05-21T09:00:00Z"
  }
}
```

Поля:

- `title` - обязательное поле.
- `description` - описание задачи.
- `status` - один из вариантов: `new`, `in_progress`, `done`. При создании пустой статус заменяется на `new`.
- `recurrence` - настройки повторяемости.

## Типы повторяемости

Поле `recurrence.type` поддерживает значения:

| Тип | Дополнительные поля | Описание |
| --- | --- | --- |
| `none` | нет | Задача без повторения |
| `every_n_days` | `days_count`, `start_date`, `end_date` | Повторять каждые N дней |
| `monthly_on_day` | `day_of_month`, `start_date`, `end_date` | Повторять каждый месяц в указанный день |
| `once_on_date` | `start_date` | Одно выполнение в указанную дату |
| `even_days` | `start_date`, `end_date` | Выполнять по четным дням месяца |
| `odd_days` | `start_date`, `end_date` | Выполнять по нечетным дням месяца |

`start_date` и `end_date` передаются в формате RFC3339, например:

```text
2026-04-21T09:00:00Z
```

Если `start_date` не передан для типов, где он не обязателен, сервис использует текущее время.

## Примеры запросов

Создать задачу с повторением каждые 3 дня:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Позвонить клиенту",
    "description": "Уточнить статус договора",
    "status": "new",
    "recurrence": {
      "type": "every_n_days",
      "days_count": 3,
      "start_date": "2026-04-21T09:00:00Z"
    }
  }'
```

Получить список задач:

```bash
curl http://localhost:8080/api/v1/tasks
```

Получить задачу по ID:

```bash
curl http://localhost:8080/api/v1/tasks/1
```

Обновить задачу:

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Позвонить клиенту",
    "description": "Клиент попросил перенести звонок",
    "status": "in_progress",
    "recurrence": {
      "type": "every_n_days",
      "days_count": 5,
      "start_date": "2026-04-22T09:00:00Z"
    }
  }'
```

Рассчитать ближайшие даты выполнения:

```bash
curl "http://localhost:8080/api/v1/tasks/1/upcoming-dates?count=5"
```

Пример ответа:

```json
[
  {
    "date": "2026-04-24T09:00:00Z"
  },
  {
    "date": "2026-04-27T09:00:00Z",
    "warning": "falls on non-working day, suggested alternatives: 2026-04-24, 2026-04-23, 2026-04-22, 2026-04-28, 2026-04-29",
    "alternative_dates": [
      "2026-04-24T09:00:00Z",
      "2026-04-23T09:00:00Z",
      "2026-04-22T09:00:00Z",
      "2026-04-28T09:00:00Z",
      "2026-04-29T09:00:00Z"
    ]
  }
]
```

Сохранить выбранные даты выполнения:

```bash
curl -X POST http://localhost:8080/api/v1/tasks/1/upcoming-dates \
  -H "Content-Type: application/json" \
  -d '[
    { "date": "2026-04-24T09:00:00Z" },
    { "date": "2026-04-28T09:00:00Z" }
  ]'
```

Удалить задачу:

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1
```

## Нерабочие дни

Сервис проверяет `start_date` на выходные и российские праздничные дни. Если дата создания или обновления задачи попадает на нерабочий день, API возвращает `400 Bad Request` с датой конфликта и альтернативными рабочими датами.

Пример ответа:

```json
{
  "error": "start_date 2026-05-09 09:00 falls on a non-working day (weekend/holiday). Please choose another date. Suggestions: 2026-05-08, 2026-05-07, 2026-05-06, 2026-05-11, 2026-05-12",
  "conflict_date": "2026-05-09 09:00",
  "suggested_dates": [
    "2026-05-08",
    "2026-05-07",
    "2026-05-06",
    "2026-05-11",
    "2026-05-12"
  ]
}
```

## Структура проекта

```text
.
├── cmd/api/                         # Точка входа приложения
├── internal/domain/task/             # Доменные модели и логика повторяемости
├── internal/usecase/task/            # Usecase-слой
├── internal/repository/postgres/     # Репозиторий PostgreSQL
├── internal/infrastructure/postgres/ # Подключение к PostgreSQL
├── internal/transport/http/          # HTTP-роутер, handlers, DTO и Swagger
├── migrations/                       # SQL-инициализация базы
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Локальный запуск без Docker для приложения

Сначала поднимите PostgreSQL:

```bash
docker compose up postgres
```

Затем запустите приложение локально:

```bash
DATABASE_DSN="postgres://postgres:postgres@localhost:5435/taskservice?sslmode=disable" \
HTTP_ADDR=":8080" \
go run ./cmd/api
```

## Проверка

Запуск тестов:

```bash
go test ./...
```

Проверка сборки:

```bash
go build ./cmd/api
```
