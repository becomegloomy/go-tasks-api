# go-tasks-api

REST API для трекера задач на Go. Учебный pet-проект: стандартная библиотека
`net/http` (роутинг Go 1.22), SQLite через `database/sql` и pure-Go драйвер
`modernc.org/sqlite` (без CGO), юнит-тесты через `httptest`.

## Стек

- **Язык:** Go 1.22+
- **HTTP:** стандартная библиотека `net/http`, роутинг через `ServeMux` с паттернами (`GET /tasks/{id}`)
- **БД:** SQLite, драйвер `modernc.org/sqlite` (чистый Go, без CGO)
- **Тесты:** пакет `testing`, `net/http/httptest`
- **CI:** GitHub Actions (`go vet`, `go test -race`)

## Эндпоинты

| Метод  | Путь          | Описание        | Коды ответа          |
|--------|---------------|-----------------|----------------------|
| GET    | `/tasks`      | Список задач    | 200                  |
| POST   | `/tasks`      | Создать задачу  | 201, 400             |
| GET    | `/tasks/{id}` | Получить задачу | 200, 400, 404        |
| PUT    | `/tasks/{id}` | Обновить задачу | 200, 400, 404        |
| DELETE | `/tasks/{id}` | Удалить задачу  | 204, 400, 404        |

### Примеры запросов

```bash
# создать задачу
curl -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"купить хлеб"}'

# список всех задач
curl localhost:8080/tasks

# получить задачу по id
curl localhost:8080/tasks/1

# обновить задачу
curl -X PUT localhost:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"title":"купить хлеб","done":true}'

# удалить задачу
curl -X DELETE localhost:8080/tasks/1
```

## Архитектура

```
go-tasks-api/
├── .github/
│   └── workflows/
│       └── ci.yml
├── internal/
│   ├── handler/
│   │   ├── task.go
│   │   └── task_test.go
│   ├── model/
│   │   └── task.go
│   └── storage/
│       └── sqlite.go
├── go.mod
├── main.go
└── README.md
```

- `internal/model` — модель `Task`
- `internal/storage` — интерфейс `Store` + реализация на SQLite
- `internal/handler` — HTTP-хендлеры
- `main.go` — сборка зависимостей, роутинг, graceful shutdown

Слой хранения скрыт за интерфейсом `Store`. В тестах используется
in-memory-реализация на `map`, в проде — SQLite. Хендлеры не знают,
откуда берутся данные, — это позволяет тестировать их без БД.

## Запуск

```bash
# подтянуть зависимости
go mod tidy

# запустить сервис (по умолчанию :8080, база в ./tasks.db)
go run .

# или с явным путём к базе
DB_PATH=/tmp/tasks.db go run .
```

Сервис поднимается на `http://localhost:8080`.

## Тесты

```bash
go test -race -cover ./...
```

Покрыты сценарии: создание задачи, получение по id, обработка
невалидного тела запроса (400), отсутствующая задача (404), удаление.

## CI

На каждый `push` в `main` и на `pull request` GitHub Actions выполняет:

```bash
go vet ./...
go test -race -cover ./...
```

Конфигурация — в `.github/workflows/ci.yml`.

## Что демонстрирует проект

- Работа с `net/http` на уровне стандартной библиотеки без фреймворков
- Роутинг через `ServeMux` с path-параметрами (Go 1.22)
- Работа с SQL через `database/sql`: `QueryContext`, `QueryRowContext`,
  `ExecContext`, обработка `sql.ErrNoRows`
- Dependency injection: хендлеры принимают интерфейс `Store`
- Юнит-тесты хендлеров через `httptest` без реальной БД
- Graceful shutdown по `SIGINT`/`SIGTERM`
- CI с `go vet` и `-race`
