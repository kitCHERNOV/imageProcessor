# Image Processing Service

Асинхронный микросервис для обработки изображений с использованием Apache Kafka и паттерна Claim Check.

## Описание

Сервис позволяет загружать изображения и обрабатывать их асинхронно, не блокируя клиента. После загрузки клиент получает ID задачи и может проверить статус обработки в любой момент.

## Функциональность

- ✅ Загрузка изображений (JPG, PNG, GIF)
- ✅ Асинхронная обработка через Kafka
- ✅ Изменение размера изображения (Resize)
- ✅ Создание миниатюр (Miniature)
- ✅ Добавление водяного знака (Watermark)
- ✅ Отслеживание статуса обработки
- ✅ Удаление обработанных изображений
- ✅ REST API с JSON responses

## Стек технологий

**Backend:**
- Go 1.25
- Chi Router (HTTP routing)
- Apache Kafka (message broker)
- SQLite (metadata storage)

**Библиотеки:**
- `github.com/IBM/sarama` — Kafka client
- `github.com/disintegration/imaging` — обработка изображений
- `github.com/fogleman/gg` — рисование водяных знаков

**Инфраструктура:**
- Docker & Docker Compose
- Kafka + Zookeeper
- Kafka UI (для мониторинга)

## Архитектура

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ POST /upload
       ▼
┌─────────────────────────────────┐
│      REST API (Chi Router)      │  ← 202 Accepted + Image ID
│                                 │
│  UploadImage Handler            │
│  ├── File validation            │
│  ├── Save to /uploads           │
│  └── Send to Kafka              │
└──────────────┬──────────────────┘
               │
               ▼
         ┌──────────────┐
         │    Kafka     │
         │   Broker     │
         └──────┬───────┘
                │
       ┌────────┴────────┐
       │                 │
       ▼                 ▼
    Consumer 1      Consumer 2
      (Part. 0)       (Part. 1)
       │                 │
       ├─ Resize    ├─ Watermark
       ├─ Miniature └─ Delete
       └─ Delete
       │                 │
       └────────┬────────┘
                ▼
         ┌──────────────┐
         │   SQLite DB  │
         │              │
         │ images table │
         └──────────────┘
                ▲
                │
         GET /image/{id}
         ← 202 Processing
         ← 200 + Image (base64)
```

**Паттерн Claim Check:**
Клиент не отправляет большие файлы в сообщение Kafka. Вместо этого:
1. Файл сохраняется на диск
2. В Kafka отправляется только ID и тип действия
3. Consumer находит файл по ID и обрабатывает его
4. Результат хранится в том же месте

## Требования

- Docker & Docker Compose
- Go 1.25+ (для локальной разработки)
- 4GB RAM
- Linux/MacOS/Windows (with WSL2)

## Быстрый старт

### 1. Клонирование и подготовка

```bash
git clone <https://github.com/kitCHERNOV/imageProcessor.git>
cd imageProcessor
```

### 2. Запуск с Docker Compose

```bash
docker-compose -f build/docker-compose.yaml up -d
```

Проверьте статус:
```bash
docker ps
```

Должны быть запущены: `zookeeper`, `kafka`, `kafka-ui`

### 3. Инициализация БД

```bash
sqlite3 storage/storage.db < migrations/schema.sql
```

### 4. Запуск приложения

**Вариант 1: Локально (Go)**
```bash
export CONFIG_PATH=config/local.yaml
go run cmd/imageProcessor/service.go
```

**Вариант 2: Docker**
Раскомментируйте секцию `app` в `docker-compose.yaml` и пересоздайте контейнеры.

### 5. Проверка статуса

- API сервис: http://localhost:8081
- Kafka UI: http://localhost:8080
- Frontend: http://localhost:3000

## API документация

### Загрузка изображения

```http
POST /upload
Content-Type: multipart/form-data

image: <file>
action: resize|miniature|watermark
```

**Response (202 Accepted):**
```json
{
  "status": "Accepted",
  "message": "image is uploaded seccessuly to do - resize",
  "image_id": 1,
  "action": "resize"
}
```

### Получение результата

```http
GET /image/{id}
```

**Response (202 Processing):**
```json
{
  "status": "pending",
  "message": "Server is handling an image",
  "image_id": 1,
  "action": "resize"
}
```

**Response (200 OK):**
```json
{
  "status": "OK",
  "image": "iVBORw0KGgoAAAANSUhEUgAAAAUA...",
  "message": "Imaged was resized"
}
```

### Удаление изображения

```http
DELETE /image/{id}
```

**Response (200 OK):**
```
image deleted
```

## Примеры использования

### cURL

```bash
# Загрузка
curl -X POST http://localhost:8081/upload \
  -F "image=@photo.jpg" \
  -F "action=resize"

# Получение результата
curl http://localhost:8081/image/1

# Удаление
curl -X DELETE http://localhost:8081/image/1
```

### JavaScript/Fetch

```javascript
// Загрузка
const formData = new FormData();
formData.append('image', fileInput.files[0]);
formData.append('action', 'resize');

const response = await fetch('http://localhost:8081/upload', {
  method: 'POST',
  body: formData
});

const data = await response.json();
const imageId = data.image_id;

// Проверка статуса
const result = await fetch(`http://localhost:8081/image/${imageId}`);
const resultData = await result.json();
```

## Структура проекта

```
.
├── build/
│   └── docker-compose.yaml       # Docker конфигурация
├── cmd/
│   └── imageProcessor/
│       └── service.go            # Точка входа приложения
├── client/
│   ├── index.html               # Web интерфейс
│   ├── main.go                  # Простой HTTP сервер для фронта
│   └── style.css                # Включен в HTML
├── config/
│   └── local.yaml               # Конфигурация приложения
├── internal/
│   ├── config/                  # Парсинг конфигурации
│   ├── handlers/                # HTTP handlers
│   ├── img-storage/             # Обработка изображений
│   │   ├── resize.go            # Изменение размера
│   │   └── watemark.go          # Водяные знаки
│   ├── kafka/                   # Kafka producer/consumer
│   ├── models/                  # Data models
│   └── storage/
│       └── sqlite/              # База данных
├── migrations/
│   └── schema.sql               # DDL для БД
├── uploads/                     # Хранилище обработанных изображений
├── go.mod & go.sum             # Зависимости Go
├── Dockerfile                   # Docker образ
└── README.md
```

## Конфигурация

**config/local.yaml:**
```yaml
storage:
  path: "storage/storage.db"      # Путь к SQLite БД
img_storage:
  path: "./uploads"               # Хранилище изображений
brokers:
  - "localhost:9092"              # Kafka brokers
```

Для production используйте переменные окружения:
```bash
export STORAGE_PATH=storage/storage.db
export IMG_STORAGE_PATH=./uploads
```

## Статусы обработки

- `pending` — ждет обработки в очереди
- `processing` — обрабатывается Consumer'ом
- `modified` — успешно обработано
- `deleted` — помечено на удаление
- `failed` — ошибка обработки (TODO)

## Мониторинг

### Проверка Kafka через UI

1. Откройте http://localhost:8080
2. Перейдите на вкладку Topics
3. Выберите `image-upload`
4. Смотрите сообщения в реальном времени

### Просмотр логов

```bash
# Backend
docker logs -f <container_id>

# Kafka
docker logs -f kafka

# Zookeeper
docker logs -f zookeeper
```

## Потенциальные улучшения (TODO)

- [ ] Добавить обработку ошибок в Consumer
- [ ] Реализовать worker pool для параллельной обработки
- [ ] Добавить retry logic для Kafka
- [ ] Параметризовать размеры изображений (вместо hardcoded)
- [ ] Добавить метрики (Prometheus)
- [ ] Реализовать cleanup старых файлов
- [ ] Добавить юнит-тесты для handlers
- [ ] Валидация расширений файлов на стороне сервера
- [ ] Компрессия изображений
- [ ] Поддержка HTTPS

## Возможные проблемы и решения

### Kafka не подключается
```bash
# Проверить статус контейнера
docker ps | grep kafka

# Перезагрузить Kafka
docker restart kafka

# Проверить логи
docker logs kafka
```

### Базе данных не существует
```bash
# Убедитесь, что таблица создана
sqlite3 storage/storage.db < migrations/schema.sql

# Или создать вручную
sqlite3 storage/storage.db
> .read migrations/schema.sql
```

### Очень медленная обработка
- Проверить нагрузку на систему
- Увеличить количество Consumer partitions в `service.go`
- Проверить размер изображений (текущий лимит 10MB)

## Контакты и вопросы

Для вопросов по архитектуре или использованию обращайтесь в issues.
