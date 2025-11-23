# Примеры использования API

## Базовый URL

```
http://localhost:8080/api/v1
```

## 1. Работа с командами

### 1.1. Создание команды

```bash
curl -X POST http://localhost:8080/api/v1/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {
        "user_id": "alice",
        "username": "Alice Smith",
        "is_active": true
      },
      {
        "user_id": "bob",
        "username": "Bob Johnson",
        "is_active": true
      },
      {
        "user_id": "charlie",
        "username": "Charlie Brown",
        "is_active": true
      }
    ]
  }'
```

**Ответ:**
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {
        "user_id": "alice",
        "username": "Alice Smith",
        "is_active": true
      },
      {
        "user_id": "bob",
        "username": "Bob Johnson",
        "is_active": true
      },
      {
        "user_id": "charlie",
        "username": "Charlie Brown",
        "is_active": true
      }
    ]
  }
}
```

### 1.2. Получение информации о команде

```bash
curl http://localhost:8080/api/v1/team/get?team_name=backend
```

**Ответ:**
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {
        "user_id": "alice",
        "username": "Alice Smith",
        "is_active": true
      },
      {
        "user_id": "bob",
        "username": "Bob Johnson",
        "is_active": true
      },
      {
        "user_id": "charlie",
        "username": "Charlie Brown",
        "is_active": true
      }
    ]
  }
}
```

## 2. Управление пользователями

### 2.1. Изменение активности пользователя

```bash
curl -X POST http://localhost:8080/api/v1/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "bob",
    "is_active": false
  }'
```

**Ответ:**
```json
{
  "user": {
    "user_id": "bob",
    "username": "Bob Johnson",
    "team_name": "backend",
    "is_active": false
  }
}
```

### 2.2. Получение PR пользователя

```bash
curl http://localhost:8080/api/v1/users/getReview?user_id=alice
```

**Ответ:**
```json
{
  "user_id": "alice",
  "pull_requests": [
    {
      "pull_request_id": "pr-1001",
      "pull_request_name": "Add authentication",
      "author_id": "bob",
      "status": "OPEN"
    },
    {
      "pull_request_id": "pr-1002",
      "pull_request_name": "Fix bug in login",
      "author_id": "charlie",
      "status": "MERGED"
    }
  ]
}
```

### 2.3. Массовая деактивация команды

```bash
curl -X POST http://localhost:8080/api/v1/users/deactivateTeam \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend"
  }'
```

**Ответ:**
```json
{
  "team_name": "backend",
  "affected_prs": [
    {
      "pull_request_id": "pr-1003",
      "pull_request_name": "Refactor API",
      "author_id": "david",
      "status": "OPEN",
      "assigned_reviewers": ["david"]
    }
  ],
  "message": "Team members deactivated successfully"
}
```

## 3. Работа с Pull Requests

### 3.1. Создание PR

```bash
curl -X POST http://localhost:8080/api/v1/pullRequests/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "alice"
  }'
```

**Ответ:**
```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "alice",
    "status": "OPEN",
    "assigned_reviewers": ["bob", "charlie"],
    "createdAt": "2025-11-23T10:30:00Z",
    "mergedAt": null
  }
}
```

### 3.2. Merge PR

```bash
curl -X POST http://localhost:8080/api/v1/pullRequests/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
```

**Ответ:**
```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "alice",
    "status": "MERGED",
    "assigned_reviewers": ["bob", "charlie"],
    "createdAt": "2025-11-23T10:30:00Z",
    "mergedAt": "2025-11-23T12:45:00Z"
  }
}
```

### 3.3. Переназначение ревьювера

```bash
curl -X POST http://localhost:8080/api/v1/pullRequests/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "bob"
  }'
```

**Ответ:**
```json
{
  "pr": {
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add authentication",
    "author_id": "alice",
    "status": "OPEN",
    "assigned_reviewers": ["charlie", "david"],
    "createdAt": "2025-11-23T10:30:00Z",
    "mergedAt": null
  },
  "replaced_by": "david"
}
```

## 4. Статистика

### 4.1. Получение полной статистики

```bash
curl http://localhost:8080/api/v1/statistics
```

**Ответ:**
```json
{
  "assignments_by_user": {
    "Alice Smith": 15,
    "Bob Johnson": 12,
    "Charlie Brown": 18,
    "David Wilson": 8
  },
  "pull_requests": {
    "total_prs": 45,
    "open_prs": 12,
    "merged_prs": 33
  }
}
```

## 5. Сценарии использования

### Сценарий 1: Создание команды и PR

```bash
# 1. Создаем команду
curl -X POST http://localhost:8080/api/v1/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "frontend",
    "members": [
      {"user_id": "eve", "username": "Eve Adams", "is_active": true},
      {"user_id": "frank", "username": "Frank Miller", "is_active": true},
      {"user_id": "grace", "username": "Grace Lee", "is_active": true}
    ]
  }'

# 2. Создаем PR от Eve
curl -X POST http://localhost:8080/api/v1/pullRequests/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-2001",
    "pull_request_name": "Redesign UI",
    "author_id": "eve"
  }'

# 3. Проверяем назначенных ревьюверов (frank и grace)
curl http://localhost:8080/api/v1/users/getReview?user_id=frank
```

### Сценарий 2: Переназначение из-за отпуска

```bash
# 1. Сотрудник уходит в отпуск
curl -X POST http://localhost:8080/api/v1/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "frank",
    "is_active": false
  }'

# 2. Переназначаем его PR на другого
curl -X POST http://localhost:8080/api/v1/pullRequests/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-2001",
    "old_user_id": "frank"
  }'
```

### Сценарий 3: Merge и закрытие PR

```bash
# 1. PR одобрен, мержим
curl -X POST http://localhost:8080/api/v1/pullRequests/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-2001"
  }'

# 2. Попытка переназначить после merge (вернет ошибку)
curl -X POST http://localhost:8080/api/v1/pullRequests/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-2001",
    "old_user_id": "grace"
  }'
# Ответ: {"error": {"code": "PR_MERGED", "message": "cannot modify merged pull request"}}
```

## 6. Обработка ошибок

### Команда уже существует

```bash
curl -X POST http://localhost:8080/api/v1/team/add \
  -H "Content-Type: application/json" \
  -d '{"team_name": "backend", "members": [...]}'
```

**Ответ (400):**
```json
{
  "error": {
    "code": "TEAM_EXISTS",
    "message": "team backend already exists"
  }
}
```

### PR не найден

```bash
curl -X POST http://localhost:8080/api/v1/pullRequests/merge \
  -H "Content-Type: application/json" \
  -d '{"pull_request_id": "non-existent"}'
```

**Ответ (404):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "pull request not found"
  }
}
```

### Нет доступных кандидатов для замены

```bash
curl -X POST http://localhost:8080/api/v1/pullRequests/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "only-active-reviewer"
  }'
```

**Ответ (409):**
```json
{
  "error": {
    "code": "NO_CANDIDATE",
    "message": "no active replacement candidate available"
  }
}
```
