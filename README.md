## loglint – кастомный Go‑линтер для логов

**Назначение**: проверка сообщений логов (`log/slog`, `go.uber.org/zap`) на:

- **строчную букву в начале** сообщения;
- **только английские буквы** (запрещаем кириллицу и другие алфавиты в тексте сообщения);
- **отсутствие спецсимволов и эмодзи**;
- **отсутствие потенциально чувствительных данных** (по ключевым словам: `password`, `token`, `secret`, `email`, и т.п.).

### Структура проекта

- `analyzer/`
  - `analyzer.go` – реализация `analysis.Analyzer` (`loglint`);
  - `analyzer_test.go` + `testdata/` – тесты на базе `analysistest`.
- `cmd/loglint/` – одиночный бинарь на основе `singlechecker` (удобно запускать через `go vet -vettool`).
- `plugin/` – точка входа для плагина golangci-lint (Go Plugin System).

### Установка и запуск как отдельного инструмента

```bash
cd loglint
go install ./cmd/loglint
```

Дальше можно использовать его в связке с `go vet`:

```bash
loglint -h           # посмотреть доступные флаги
loglint ./...        # запустить анализатор напрямую
```

Или как `vettool`:

```bash
go vet -vettool=$(which loglint) ./...
```

### Тесты

Запуск всех тестов:

```bash
cd loglint
go test ./...

### Интеграция с golangci-lint (Go Plugin System)

1. **Собрать плагин**:

```bash
cd loglint
go build -buildmode=plugin -o loglint.so ./plugin
```

2. **Подключить в `.golangci.yml`** вашего проекта:

```yaml
version: "2"

linters:
  enable:
    - loglint

  settings:
    custom:
      loglint:
        path: ./loglint.so
        description: Custom log message linter (slog + zap)
        original-url: loglint
```

3. **Запустить golangci-lint**:

```bash
golangci-lint run ./...
```

### Поддерживаемые логгеры и что именно проверяется

- **`log/slog`**:
  - `slog.Info("message", ...)`;
  - методы `Logger.Info(...)`, `Logger.Error(...)` и т.п.
- **`go.uber.org/zap`**:
  - методы `(*zap.Logger).Info`, `Debug`, `Warn`, `Error` (и `*f`‑варианты).

Проверяется **первый строковый аргумент** вызова логгера, если это строковый литерал.


### CI/CD

- **GitHub Actions**: файл конфигурации находится в `.github/workflows/ci.yml`.
  - при каждом `push` в ветки `main`/`master` и при `pull_request`:
    - поднимается runner с Ubuntu;
    - ставится Go `1.25.6`;
    - выполняется `go test ./...` в каталоге `loglint`.

- **GitLab CI**: файл конфигурации `.gitlab-ci.yml` в корне репозитория.
  - для каждого пуша:
    - запускается job `go_test` в образе `golang:1.25`;
    - выполняется `cd loglint && go test ./...`.