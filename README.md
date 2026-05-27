# SFEDU CRM — Спортивный зал

Backend CRM-система для управления спортивным залом на Go.

## Стек

- **Go 1.22** — net/http ServeMux (без фреймворков)
- **PostgreSQL 16** — основная БД
- **pgx/v5** — драйвер PostgreSQL
- **golang-migrate** — миграции
- **bcrypt** — хеширование паролей
- **golang-jwt/jwt/v5** — JWT аутентификация
- **Docker + Docker Compose** — деплой

---

## Быстрый старт

```bash
# 1. Скопировать конфиг
cp .env.example .env
# Отредактировать .env: выставить JWT_SECRET

# 2. Запустить через Docker
make docker-up

# 3. Или локально (PostgreSQL должен быть запущен)
go run ./cmd/server/...
```

---

## Архитектура каталогов

```
sfedu-crm/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа, DI-сборка
├── internal/
│   ├── config/                  # Конфигурация из env
│   ├── domain/                  # Сущности и ошибки
│   │   ├── user.go
│   │   ├── trainer.go
│   │   ├── training.go
│   │   ├── shop.go
│   │   ├── finance.go
│   │   └── errors.go
│   ├── repository/
│   │   ├── interfaces.go        # Интерфейсы репозиториев
│   │   └── postgres/            # Реализации на pgx
│   ├── service/                 # Бизнес-логика
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── schedule.go
│   │   ├── finance_shop.go
│   │   ├── trainer.go
│   │   └── ai_assistant.go      # Заглушка ИИ-ассистента
│   ├── handler/http/            # HTTP handlers (net/http)
│   │   ├── router.go            # Все маршруты
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── schedule.go
│   │   ├── finance_shop.go
│   │   ├── trainer_ai.go
│   │   └── helpers.go
│   ├── middleware/              # Auth, CORS, Logging
│   └── validator/               # Валидация (ФИО, телефон, email)
├── pkg/
│   ├── jwt/                     # JWT Manager
│   ├── hash/                    # bcrypt wrapper
│   └── logger/                  # slog logger
├── migrations/                  # SQL миграции
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## Роли и права

| Действие                          | Клиент | Менеджер | Админ |
|-----------------------------------|:------:|:--------:|:-----:|
| Просмотр своего профиля           | ✅     | ✅       | ✅    |
| Редактирование своего профиля     | ✅     | ✅       | ✅    |
| Смена своего пароля               | ✅     | ✅       | ✅    |
| Просмотр своего расписания        | ✅     | ✅       | ✅    |
| Подача заявки на тренировку       | ✅     | ✅       | ✅    |
| Список всех клиентов              | ❌     | ✅       | ✅    |
| Создание пользователей            | ❌     | ✅       | ✅    |
| Редактирование данных клиентов    | ❌     | ✅       | ✅    |
| Принятие заявок на тренировку     | ❌     | ✅       | ✅    |
| Перенос занятий (дата/время/тренер)| ❌    | ✅       | ✅    |
| Управление товарами в магазине    | ❌     | ✅       | ✅    |
| Пополнение баланса клиентов       | ❌     | ✅       | ✅    |
| Удаление пользователей            | ❌     | ❌       | ✅    |
| Блокировка/разблокировка юзеров   | ❌     | ❌       | ✅    |
| Удаление занятий                  | ❌     | ❌       | ✅    |
| Управление тренерами              | ❌     | ❌       | ✅    |
| Просмотр всех финансов            | ❌     | ❌       | ✅    |
| Финансовая сводка                 | ❌     | ❌       | ✅    |

---

## API Endpoints

### Публичные (без токена)

```
POST /api/v1/auth/apply          Подать заявку на регистрацию (форма «Записаться»)
POST /api/v1/auth/login          Войти (email + пароль) → JWT токен
GET  /api/v1/trainers            Тренерский состав
GET  /api/v1/trainers/{id}       Профиль тренера
```

### Профиль (все авторизованные)

```
GET  /api/v1/users/me            Мой профиль
PUT  /api/v1/users/me            Обновить профиль (ФИО, телефон, email, пол)
PUT  /api/v1/users/me/password   Сменить пароль
```

### Расписание

```
GET  /api/v1/schedule                          Расписание (фильтры: date_from, date_to)
GET  /api/v1/schedule/{id}                     Детали занятия
POST /api/v1/schedule/requests                 Подать заявку на тренировку [клиент]
GET  /api/v1/schedule/requests/my             Мои заявки [клиент]
GET  /api/v1/schedule/requests                 Все pending-заявки [менеджер/админ]
POST /api/v1/schedule/requests/{id}/approve    Принять заявку [менеджер/админ]
POST /api/v1/schedule/requests/{id}/reject     Отклонить заявку [менеджер/админ]
PUT  /api/v1/schedule/{id}                     Перенести занятие [менеджер/админ]
DELETE /api/v1/schedule/{id}                   Удалить занятие [админ]
```

### Клиенты [менеджер/админ]

```
GET    /api/v1/users              Список клиентов (фильтры: role, search)
POST   /api/v1/users              Создать пользователя
GET    /api/v1/users/{id}         Профиль пользователя
PUT    /api/v1/users/{id}         Редактировать данные
DELETE /api/v1/users/{id}         Удалить [только админ]
PATCH  /api/v1/users/{id}/activate    Активировать [только админ]
PATCH  /api/v1/users/{id}/deactivate  Деактивировать [только админ]
```

### Заявки на регистрацию [менеджер/админ]

```
GET  /api/v1/applications                 Список заявок (?status=pending)
POST /api/v1/applications/{id}/approve    Принять (+ задать пароль)
POST /api/v1/applications/{id}/reject     Отклонить
```

### Финансы

```
GET  /api/v1/finance/me/payments   Моя история платежей [клиент]
POST /api/v1/finance/me/topup      Пополнить свой баланс [клиент]
POST /api/v1/finance/topup         Пополнить баланс клиента [менеджер/админ]
GET  /api/v1/finance/payments      Все платежи [только админ]
GET  /api/v1/finance/summary       Финансовая сводка [только админ]
```

### Магазин

```
GET  /api/v1/shop/products           Каталог (?category=subscription|sports)
GET  /api/v1/shop/products/{id}      Детали товара
POST /api/v1/shop/purchase           Купить товар/абонемент [клиент]
GET  /api/v1/shop/my-subscriptions   Мои абонементы [клиент]
POST /api/v1/shop/products           Добавить товар [менеджер/админ]
PUT  /api/v1/shop/products/{id}      Обновить товар [менеджер/админ]
DELETE /api/v1/shop/products/{id}    Удалить товар [только админ]
```

### Тренеры (управление) [только админ]

```
POST   /api/v1/trainers           Добавить тренера
PUT    /api/v1/trainers/{id}      Обновить профиль тренера
DELETE /api/v1/trainers/{id}      Удалить тренера
```

### ИИ Спортсмен

```
POST /api/v1/ai/chat    Диалог с ИИ-ассистентом (заглушка, готова к интеграции)
```

---

## Валидация

- **ФИО** — минимум 2 слова, только буквы
- **Телефон** — российский формат: `+7(987)654-32-10`, `+79871234567`, `89871234567`
- **Email** — только `@gmail.com` и `@mail.ru`
- **Пароль** — минимум 8 символов
