Проект содержит:
- API на Golang (порт 8080)
- Python gRPC сервер (порт 4041)
- PostgreSQL (порт 5432)
- pgAdmin (порт 5050)
- Миграции для БД
- API для Postman API.postman_collection.json
- Контракт .proto для gRPC сервера

Везде пароль **passw0rd**

## Запуск
```bash
git clone https://github.com/liriquew/control_system.git
cd control_system
docker-compose up --build
# или
docker compose up --build
