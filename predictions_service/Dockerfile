# Dockerfile for Python server
FROM python:3.12

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем все файлы в контейнер
COPY . .

# Устанавливаем зависимости
RUN pip install --no-cache-dir -r requirements.txt

# Открываем порт (например, для gRPC 4041)
EXPOSE 4041

# Запуск сервера
CMD ["python", "main.py"]
